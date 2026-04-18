package appserviceimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	secretDefaultFileUID  = "0"
	secretDefaultFileGID  = "0"
	secretDefaultFileMode = 444
)

func (s *service) CreateSwarmSecret(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secret *entity.Secret,
) (err error) {
	swarmRef := secret.SwarmRef
	if swarmRef == nil || swarmRef.File == nil {
		return nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(secret.Key))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, secretDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, secretDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, secretDefaultFileMode)

	// Create the secret in docker swarm
	secretName := app.LocalKey + "_" + strings.ToLower(secret.Key)
	secretVal := reflectutil.UnsafeStrToBytes(secret.Value.MustGetPlain())
	secretResp, err := s.dockerManager.SecretCreate(ctx, secretName, secretVal, func(sec *swarm.SecretSpec) {
		sec.Labels = map[string]string{
			docker.StackLabelNamespace: app.Project.Key,
		}
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrInfraConflict) || errors.Is(err, apperrors.ErrInfraAlreadyExists) {
			// Delete the orphan secret, then retry this action
			if err := s.deleteOrphanSwarmSecret(ctx, db, app, secretName); err == nil {
				return s.CreateSwarmSecret(ctx, db, app, secret)
			}
		}
		return apperrors.Wrap(err)
	}
	swarmRef.SecretID = secretResp.ID
	swarmRef.SecretName = secretName

	// Add the secret to the swarm service of the app
	err = s.addSwarmSecretToService(ctx, app.ServiceID, swarmRef)
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
			err = s.addSwarmSecretToService(ctx, childApp.ServiceID, swarmRef)
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
		inheritedSecretSetting, err := s.settingRepo.GetByName(ctx, db, base.NewSettingScopeApp(app.ParentID, app.ProjectID),
			base.SettingTypeSecret, secret.Key, false)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if inheritedSecretSetting != nil {
			err = s.addSwarmSecretToService(ctx, app.ServiceID, inheritedSecretSetting.MustAsSecret().SwarmRef)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the secret
	err = s.dockerManager.SecretRemove(ctx, secret.SwarmRef.SecretID)
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
	err = s.CreateSwarmSecret(ctx, db, app, newSecret)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) addSwarmSecretToService(
	ctx context.Context,
	serviceID string,
	swarmRef *entity.SwarmSecretRef,
) (err error) {
	if serviceID == "" || swarmRef == nil || swarmRef.SecretID == "" {
		return nil
	}

	swarmSvc, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Only add the secret to the swarm service when the target file name is not used
	// by another secret.
	for _, sec := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets {
		if sec.File != nil && sec.File.Name == swarmRef.File.Name {
			return nil
		}
	}

	swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets = append(swarmSvc.Spec.TaskTemplate.ContainerSpec.Secrets,
		&swarm.SecretReference{
			File: &swarm.SecretReferenceFileTarget{
				Name: swarmRef.File.Name,
				UID:  swarmRef.File.UID,
				GID:  swarmRef.File.GID,
				Mode: swarmRef.File.Mode.ToFileMode(),
			},
			SecretID:   swarmRef.SecretID,
			SecretName: swarmRef.SecretName,
		})

	_, err = s.dockerManager.ServiceUpdate(ctx, serviceID, &swarmSvc.Version, &swarmSvc.Spec)
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

	swarmSvc, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

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

	orphanSwarmSec, err := s.dockerManager.SecretInspect(ctx, secretNameOrID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.Wrap(err)
	}

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
	err = s.dockerManager.SecretRemove(ctx, orphanSwarmSec.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
