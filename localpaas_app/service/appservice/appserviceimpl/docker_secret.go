package appserviceimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	secretDefaultFileUID  = "0"
	secretDefaultFileGID  = "0"
	secretDefaultFileMode = 444
)

func (s *service) CreateSwarmSecrets(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secrets []*entity.Secret,
) (refs []*entity.SwarmSecretRef, err error) {
	refs = make([]*entity.SwarmSecretRef, 0, len(secrets))
	for _, secret := range secrets {
		ref, err := s.createSwarmSecret(ctx, db, app, secret)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		refs = append(refs, ref)
	}
	return refs, s.addSwarmSecretsToService(ctx, db, app, refs)
}

func (s *service) CreateSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secret *entity.Secret,
) (*entity.SwarmSecretRef, error) {
	refs, err := s.CreateSwarmSecrets(ctx, db, app, []*entity.Secret{secret})
	ref, _ := gofn.First(refs)
	return ref, apperrors.Wrap(err)
}

func (s *service) createSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secret *entity.Secret,
) (ref *entity.SwarmSecretRef, err error) {
	swarmRef := secret.SwarmRef
	if swarmRef == nil || swarmRef.File == nil {
		return nil, nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(secret.Key))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, secretDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, secretDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, secretDefaultFileMode)

	// Create the secret in docker swarm
	secretName := app.LocalKey + "_" + strings.ToLower(secret.Key)
	secretResp, err := s.dockerManager.SecretCreate(ctx, secretName, secret.ValueAsBytes(),
		func(opts *client.SecretCreateOptions) {
			opts.Spec.Labels = map[string]string{
				docker.StackLabelNamespace: app.Project.Key,
			}
		})
	if err != nil {
		if errors.Is(err, apperrors.ErrInfraConflict) || errors.Is(err, apperrors.ErrInfraAlreadyExists) {
			// Delete the orphan secret, then retry this action
			if err := s.deleteOrphanSwarmSecret(ctx, db, app, secretName); err == nil {
				return s.createSwarmSecret(ctx, db, app, secret)
			}
		}
		return nil, apperrors.Wrap(err)
	}
	swarmRef.SecretID = secretResp.ID
	swarmRef.SecretName = secretName
	return swarmRef, nil
}

func (s *service) addSwarmSecretsToService(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	refs []*entity.SwarmSecretRef,
) (err error) {
	if app.ServiceID == "" || len(refs) == 0 {
		return nil
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	swarmSvc := &inspect.Service
	containerSpec := swarmSvc.Spec.TaskTemplate.ContainerSpec

	for _, swarmRef := range refs {
		if swarmRef == nil || swarmRef.SecretID == "" {
			continue
		}
		// Only add the secret to the swarm service when the target file name is not used by another secret
		_, inUse := gofn.Find(containerSpec.Secrets, func(sec *swarm.SecretReference) bool {
			return sec.File != nil && sec.File.Name == swarmRef.File.Name
		})
		if inUse {
			continue
		}
		containerSpec.Secrets = append(containerSpec.Secrets, &swarm.SecretReference{
			File: &swarm.SecretReferenceFileTarget{
				Name: swarmRef.File.Name,
				UID:  swarmRef.File.UID,
				GID:  swarmRef.File.GID,
				Mode: swarmRef.File.Mode.ToFileMode(),
			},
			SecretID:   swarmRef.SecretID,
			SecretName: swarmRef.SecretName,
		})
	}

	_, err = s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &swarmSvc.Version, &swarmSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps
	if app.ParentID == "" {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.addSwarmSecretsToService(ctx, db, childApp, refs)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}

func (s *service) DeleteSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secret *entity.Secret,
) (err error) {
	if secret.SwarmRef == nil || secret.SwarmRef.SecretID == "" {
		return nil
	}

	// Remove the secret from the swarm service of the app
	err = s.removeSwarmSecretFromService(ctx, app.ServiceID, secret.SwarmRef)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the secret from the child apps
	if app.ParentID == "" { //nolint:nestif
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.DeleteSwarmSecret(ctx, db, childApp, secret)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	} else {
		// This is a child app, we may need to restore the inherited secret having the same name as this
		inheritedSecretSetting, err := s.settingRepo.GetByName(ctx, db, base.NewObjectScopeApp(app.ParentID, app.ProjectID),
			base.SettingTypeSecret, secret.Key, false)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if inheritedSecretSetting != nil {
			err = s.addSwarmSecretsToService(ctx, db, app,
				[]*entity.SwarmSecretRef{inheritedSecretSetting.MustAsSecret().SwarmRef})
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the secret
	_, err = s.dockerManager.SecretRemove(ctx, secret.SwarmRef.SecretID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	secret.SwarmRef.SecretID = ""
	secret.SwarmRef.SecretName = ""

	return nil
}

func (s *service) UpdateSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldSecret, newSecret *entity.Secret,
) (err error) {
	// Remove the old secret from services then delete it from the swarm
	err = s.DeleteSwarmSecret(ctx, db, app, oldSecret)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create a secret in the swarm then add it to the services
	_, err = s.CreateSwarmSecret(ctx, db, app, newSecret)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) removeSwarmSecretFromService(
	ctx context.Context,
	serviceID string,
	swarmRef *entity.SwarmSecretRef,
) (err error) {
	if serviceID == "" || swarmRef == nil || swarmRef.SecretID == "" {
		return nil
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	swarmSvc := &inspect.Service

	hasChanges := false
	updateSecrets := make([]*swarm.SecretReference, 0, len(swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets))
	for _, sec := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets {
		if swarmRef.SecretID == sec.SecretID {
			hasChanges = true
			continue
		}
		updateSecrets = append(updateSecrets, sec)
	}
	if !hasChanges {
		return nil
	}

	swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets = updateSecrets
	_, err = s.dockerManager.ServiceUpdate(ctx, serviceID, &swarmSvc.Version, &swarmSvc.Spec)
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
	if app.ParentID == "" {
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
