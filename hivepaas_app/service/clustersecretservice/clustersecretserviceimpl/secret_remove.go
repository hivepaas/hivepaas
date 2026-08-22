package clustersecretserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) SecretRemove(
	ctx context.Context,
	secretID string, // secret id in docker
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if secretID == "" {
		return nil
	}
	fn := func() error {
		_, err := s.dockerManager.SecretRemove(ctx, secretID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return nil
			}
			return apperrors.Wrap(err)
		}
		return nil
	}
	if retryMax > 0 {
		if retryDelay <= 0 {
			retryDelay = itemRemovalRetryDelay
		}
		err = gofn.ExecRetryCtx(ctx, fn, retryMax, retryDelay, gofn.ExecRetryDelayIncr(itemRemovalRetryIncr))
	} else {
		err = fn()
	}
	if err != nil {
		// TODO: create a cleanup task
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) SecretsRemove(
	ctx context.Context,
	secretIDs []string, // secret ids in docker
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if len(secretIDs) == 0 {
		return nil
	}
	if len(secretIDs) == 1 {
		return s.SecretRemove(ctx, secretIDs[0], retryMax, retryDelay)
	}
	errMap := gofn.ExecTaskFuncEx(ctx, 10, false, //nolint:mnd
		func(ctx context.Context, itemID string) error {
			return s.SecretRemove(ctx, itemID, retryMax, retryDelay)
		}, secretIDs...)
	for _, e := range errMap {
		err = errors.Join(err, e)
	}
	return err
}

func (s *service) RemoveSecretForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secrets ...*entity.Secret,
) (err error) {
	secretRefs := make([]*entity.SwarmSecretRef, 0, len(secrets))
	for _, secret := range secrets {
		if secret == nil || secret.SwarmRef == nil || secret.SwarmRef.SecretID == "" {
			continue
		}
		secretRefs = append(secretRefs, secret.SwarmRef)
	}
	if len(secretRefs) == 0 {
		return nil
	}

	// Remove the secret items from the swarm service of the app
	err = s.removeSwarmSecretFromService(ctx, app.ServiceID, secretRefs...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the secrets from the child apps
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.removeSwarmSecretFromService(ctx, childApp.ServiceID, secretRefs...)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the secret items from docker
	for _, secretRef := range secretRefs {
		err = s.SecretRemove(ctx, secretRef.SecretID, itemRemovalRetryMax, itemRemovalRetryDelay)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		secretRef.SecretID = ""
		secretRef.SecretName = ""
	}

	return nil
}

func (s *service) removeSwarmSecretFromService(
	ctx context.Context,
	serviceID string,
	swarmRefs ...*entity.SwarmSecretRef,
) (err error) {
	if serviceID == "" {
		return nil
	}
	err = s.dockerManager.ServiceUpdateFunc(ctx, serviceID, nil,
		func(_ int, swarmSvc *swarm.Service) (bool, error) {
			hasChanges := false
			updateSecrets := make([]*swarm.SecretReference, 0, len(swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets))
			for _, swarmRef := range swarmRefs {
				if swarmRef == nil || swarmRef.SecretID == "" {
					continue
				}
				for _, cfg := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets {
					if swarmRef.SecretID == cfg.SecretID {
						hasChanges = true
						continue
					}
					updateSecrets = append(updateSecrets, cfg)
				}
			}
			if !hasChanges {
				return false, nil
			}
			swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets = updateSecrets
			return true, nil
		}, itemRemovalRetryMax, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) deleteOrphanSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secretNameOrID string,
) (err error) {
	if secretNameOrID == "" {
		return nil
	}

	inspect, err := s.dockerManager.SecretInspect(ctx, secretNameOrID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.Wrap(err)
	}
	orphanSwarmSec := &inspect.Secret

	orphanSwarmRef := &entity.SwarmSecretRef{
		File:       &entity.SwarmRefFileTarget{},
		SecretID:   orphanSwarmSec.ID,
		SecretName: orphanSwarmSec.Spec.Name,
	}

	// Remove the secret from the swarm service of the app
	err = s.removeSwarmSecretFromService(ctx, app.ServiceID, orphanSwarmRef)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the secret from the child apps
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.removeSwarmSecretFromService(ctx, childApp.ServiceID, orphanSwarmRef)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the secret
	_, err = s.dockerManager.SecretRemove(ctx, orphanSwarmSec.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
