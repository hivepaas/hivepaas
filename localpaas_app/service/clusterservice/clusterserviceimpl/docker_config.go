package clusterserviceimpl

import (
	"context"
	"errors"
	"strings"
	"time"

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
	configDefaultFileUID  = "0"
	configDefaultFileGID  = "0"
	configDefaultFileMode = 444
)

func (s *service) CreateConfigsForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	configs []*entity.ConfigFile,
) (refs []*entity.SwarmConfigRef, err error) {
	refs = make([]*entity.SwarmConfigRef, 0, len(configs))
	for _, cfg := range configs {
		ref, err := s.createSwarmConfig(ctx, db, app, cfg)
		if err != nil {
			return nil, apperrors.New(err)
		}
		refs = append(refs, ref)
	}
	return refs, s.addSwarmConfigsToService(ctx, db, app, refs)
}

func (s *service) CreateConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (*entity.SwarmConfigRef, error) {
	refs, err := s.CreateConfigsForApp(ctx, db, app, []*entity.ConfigFile{config})
	ref, _ := gofn.First(refs)
	return ref, apperrors.New(err)
}

func (s *service) createSwarmConfig(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (ref *entity.SwarmConfigRef, err error) {
	swarmRef := config.SwarmRef
	if swarmRef == nil || swarmRef.File == nil {
		return nil, nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(config.Name))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, configDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, configDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, configDefaultFileMode)

	// Create the config in docker swarm
	configName := app.LocalKey + "_" + strings.ToLower(config.Name)
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
		return nil, apperrors.New(err)
	}
	swarmRef.ConfigID = configResp.ID
	swarmRef.ConfigName = configName
	return swarmRef, nil
}

func (s *service) addSwarmConfigsToService(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	refs []*entity.SwarmConfigRef,
) (err error) {
	if app.ServiceID == "" || len(refs) == 0 {
		return nil
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	swarmSvc := &inspect.Service
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

	_, err = s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &swarmSvc.Version, &swarmSvc.Spec)
	if err != nil {
		return apperrors.New(err)
	}

	// If this app is parent of some other apps
	if app.ParentID == "" {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.New(err)
		}
		for _, childApp := range childApps {
			err = s.addSwarmConfigsToService(ctx, db, childApp, refs)
			if err != nil {
				return apperrors.New(err)
			}
		}
	}

	return nil
}

func (s *service) DeleteConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (err error) {
	if config.SwarmRef == nil || config.SwarmRef.ConfigID == "" {
		return nil
	}

	// Remove the config from the swarm service of the app
	err = s.removeSwarmConfigFromService(ctx, app.ServiceID, config.SwarmRef)
	if err != nil {
		return apperrors.New(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if app.ParentID == "" { //nolint:nestif
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.New(err)
		}
		for _, childApp := range childApps {
			err = s.DeleteConfigForApp(ctx, db, childApp, config)
			if err != nil {
				return apperrors.New(err)
			}
		}
	} else {
		// This is a child app, we may need to restore the inherited config having the same name as this
		inheritedConfigSetting, err := s.settingRepo.GetByName(ctx, db, base.NewObjectScopeApp(app.ParentID, app.ProjectID),
			base.SettingTypeConfigFile, config.Name, false)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(err)
		}
		if inheritedConfigSetting != nil {
			err = s.addSwarmConfigsToService(ctx, db, app,
				[]*entity.SwarmConfigRef{inheritedConfigSetting.MustAsConfigFile().SwarmRef})
			if err != nil {
				return apperrors.New(err)
			}
		}
	}

	// Now delete the config
	_, err = s.dockerManager.ConfigRemove(ctx, config.SwarmRef.ConfigID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}
	config.SwarmRef.ConfigID = ""
	config.SwarmRef.ConfigName = ""

	return nil
}

func (s *service) UpdateConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldConfig, newConfig *entity.ConfigFile,
) (err error) {
	// Remove the old config from services then delete it from the swarm
	err = s.DeleteConfigForApp(ctx, db, app, oldConfig)
	if err != nil {
		return apperrors.New(err)
	}

	// Create a config in the swarm then add it to the services
	_, err = s.CreateConfigForApp(ctx, db, app, newConfig)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) removeSwarmConfigFromService(
	ctx context.Context,
	serviceID string,
	swarmRef *entity.SwarmConfigRef,
) (err error) {
	if serviceID == "" || swarmRef == nil || swarmRef.ConfigID == "" {
		return nil
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.New(err)
	}
	swarmSvc := &inspect.Service

	hasChanges := false
	updateConfigs := make([]*swarm.ConfigReference, 0, len(swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs))
	for _, cfg := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs {
		if swarmRef.ConfigID == cfg.ConfigID {
			hasChanges = true
			continue
		}
		updateConfigs = append(updateConfigs, cfg)
	}
	if !hasChanges {
		return nil
	}

	swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs = updateConfigs
	_, err = s.dockerManager.ServiceUpdate(ctx, serviceID, &swarmSvc.Version, &swarmSvc.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) deleteOrphanSwarmConfig(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	configNameOrID string,
) (err error) {
	if configNameOrID == "" {
		return nil
	}

	inspect, err := s.dockerManager.ConfigInspect(ctx, configNameOrID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.New(err)
	}
	orphanSwarmConfig := &inspect.Config

	orphanSwarmRef := &entity.SwarmConfigRef{
		File:       &entity.SwarmRefFileTarget{},
		ConfigID:   orphanSwarmConfig.ID,
		ConfigName: orphanSwarmConfig.Spec.Name,
	}

	// Remove the config from the swarm service of the app
	err = s.removeSwarmConfigFromService(ctx, app.ServiceID, orphanSwarmRef)
	if err != nil {
		return apperrors.New(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if app.ParentID == "" {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.New(err)
		}
		for _, childApp := range childApps {
			err = s.removeSwarmConfigFromService(ctx, childApp.ServiceID, orphanSwarmRef)
			if err != nil {
				return apperrors.New(err)
			}
		}
	}

	// Now delete the config
	_, err = s.dockerManager.ConfigRemove(ctx, orphanSwarmConfig.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) ConfigRemove(
	ctx context.Context,
	configID string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if configID == "" {
		return nil
	}
	fn := func() error {
		_, err := s.dockerManager.ConfigRemove(ctx, configID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return nil
			}
			return apperrors.New(err)
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
		return apperrors.New(err)
	}
	return nil
}

func (s *service) ConfigsRemove(
	ctx context.Context,
	configIDs []string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if len(configIDs) == 0 {
		return nil
	}
	if len(configIDs) == 1 {
		return s.ConfigRemove(ctx, configIDs[0], retryMax, retryDelay)
	}
	errMap := gofn.ExecTaskFuncEx(ctx, 10, false, //nolint:mnd
		func(ctx context.Context, itemID string) error {
			return s.ConfigRemove(ctx, itemID, retryMax, retryDelay)
		}, configIDs...)
	for _, e := range errMap {
		err = errors.Join(err, e)
	}
	return err
}
