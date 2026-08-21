package clustersecretserviceimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) CreateSecretsForApp(
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
	return refs, s.addSwarmSecretsToService(ctx, db, app, refs...)
}

func (s *service) CreateSecretForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	secret *entity.Secret,
) (*entity.SwarmSecretRef, error) {
	refs, err := s.CreateSecretsForApp(ctx, db, app, []*entity.Secret{secret})
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
	// Only create an item in docker if it is configured to mount to file system
	if swarmRef == nil || swarmRef.File == nil || swarmRef.File.Name == "" {
		return nil, nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(secret.Key))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, secretDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, secretDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, secretDefaultFileMode)

	// Create the secret in docker swarm
	secretName := app.GlobalKey + "_" + strings.ToLower(secret.Key)
	secretBytes, err := secret.ValueAsBytes()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	secretResp, err := s.dockerManager.SecretCreate(ctx, secretName, secretBytes,
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
	refs ...*entity.SwarmSecretRef,
) (err error) {
	if len(refs) == 0 || app.ServiceID == "" {
		return nil
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, swarmSvc *swarm.Service) (bool, error) {
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
			return true, nil
		}, itemRemovalRetryMax, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.addSwarmSecretsToService(ctx, db, childApp, refs...)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}
