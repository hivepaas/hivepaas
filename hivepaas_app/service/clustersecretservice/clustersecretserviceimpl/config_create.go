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

func (s *service) CreateConfigsForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	configs []*entity.ConfigFile,
) (refs []*entity.SwarmConfigRef, err error) {
	refs = make([]*entity.SwarmConfigRef, 0, len(configs))
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		ref, err := s.createSwarmConfig(ctx, db, app, cfg)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		refs = append(refs, ref)
	}
	return refs, s.addSwarmConfigsToService(ctx, db, app, refs...)
}

func (s *service) CreateConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (*entity.SwarmConfigRef, error) {
	if config == nil {
		return nil, nil
	}
	refs, err := s.CreateConfigsForApp(ctx, db, app, []*entity.ConfigFile{config})
	ref, _ := gofn.First(refs)
	return ref, apperrors.Wrap(err)
}

func (s *service) createSwarmConfig(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (ref *entity.SwarmConfigRef, err error) {
	swarmRef := config.SwarmRef
	// Only create an item in docker if it is configured to mount to file system
	if swarmRef == nil || swarmRef.File == nil || swarmRef.File.Name == "" {
		return nil, nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(config.Name))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, configDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, configDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, configDefaultFileMode)

	// Create the config in docker swarm
	configName := app.GlobalKey + "_" + strings.ToLower(config.Name)
	configResp, err := s.dockerManager.ConfigCreate(ctx, configName, config.ContentAsBytes(),
		func(opts *client.ConfigCreateOptions) {
			opts.Spec.Labels = map[string]string{
				docker.StackLabelNamespace: app.Project.Key,
			}
		})
	if err != nil {
		if errors.Is(err, apperrors.ErrInfraConflict) || errors.Is(err, apperrors.ErrInfraAlreadyExists) {
			// Delete the orphan config, then retry this action
			if err := s.deleteOrphanSwarmConfig(ctx, db, app, configName); err == nil {
				return s.createSwarmConfig(ctx, db, app, config)
			}
		}
		return nil, apperrors.Wrap(err)
	}
	swarmRef.ConfigID = configResp.ID
	swarmRef.ConfigName = configName
	return swarmRef, nil
}

func (s *service) addSwarmConfigsToService(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	refs ...*entity.SwarmConfigRef,
) (err error) {
	if len(refs) == 0 || app.ServiceID == "" {
		return nil
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, swarmSvc *swarm.Service) (bool, error) {
			containerSpec := swarmSvc.Spec.TaskTemplate.ContainerSpec
			for _, swarmRef := range refs {
				if swarmRef == nil || swarmRef.ConfigID == "" {
					continue
				}
				// Only add the config to the swarm service when the target file name is not used by another config
				_, inUse := gofn.Find(containerSpec.Configs, func(cfg *swarm.ConfigReference) bool {
					return cfg.File != nil && cfg.File.Name == swarmRef.File.Name
				})
				if inUse {
					continue
				}
				containerSpec.Configs = append(containerSpec.Configs, &swarm.ConfigReference{
					File: &swarm.ConfigReferenceFileTarget{
						Name: swarmRef.File.Name,
						UID:  swarmRef.File.UID,
						GID:  swarmRef.File.GID,
						Mode: swarmRef.File.Mode.ToFileMode(),
					},
					ConfigID:   swarmRef.ConfigID,
					ConfigName: swarmRef.ConfigName,
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
			err = s.addSwarmConfigsToService(ctx, db, childApp, refs...)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}
