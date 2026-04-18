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
	configDefaultFileUID  = "0"
	configDefaultFileGID  = "0"
	configDefaultFileMode = 444
)

func (s *service) CreateSwarmConfig(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	config *entity.ConfigFile,
) (err error) {
	swarmRef := config.SwarmRef
	if swarmRef == nil || swarmRef.File == nil {
		return nil
	}

	swarmRef.File.Name = gofn.Coalesce(swarmRef.File.Name, strings.ToLower(config.Name))
	swarmRef.File.UID = gofn.Coalesce(swarmRef.File.UID, configDefaultFileUID)
	swarmRef.File.GID = gofn.Coalesce(swarmRef.File.GID, configDefaultFileGID)
	swarmRef.File.Mode = gofn.Coalesce(swarmRef.File.Mode, configDefaultFileMode)

	// Create the config in docker swarm
	configName := app.LocalKey + "_" + strings.ToLower(config.Name)
	configVal := reflectutil.UnsafeStrToBytes(config.Content)
	configResp, err := s.dockerManager.ConfigCreate(ctx, configName, configVal, func(cfg *swarm.ConfigSpec) {
		cfg.Labels = map[string]string{
			docker.StackLabelNamespace: app.Project.Key,
		}
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrInfraConflict) || errors.Is(err, apperrors.ErrInfraAlreadyExists) {
			// Delete the orphan config, then retry this action
			if err := s.deleteOrphanSwarmConfig(ctx, db, app, configName); err == nil {
				return s.CreateSwarmConfig(ctx, db, app, config)
			}
		}
		return apperrors.Wrap(err)
	}
	swarmRef.ConfigID = configResp.ID
	swarmRef.ConfigName = configName

	// Add the config to the swarm service of the app
	err = s.addSwarmConfigToService(ctx, app.ServiceID, swarmRef)
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
			err = s.addSwarmConfigToService(ctx, childApp.ServiceID, swarmRef)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}

func (s *service) DeleteSwarmConfig(
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
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if app.ParentID == "" { //nolint:nestif
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.DeleteSwarmConfig(ctx, db, childApp, config)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	} else {
		// This is a child app, we may need to restore the inherited config having the same name as this
		inheritedConfigSetting, err := s.settingRepo.GetByName(ctx, db, base.NewSettingScopeApp(app.ParentID, app.ProjectID),
			base.SettingTypeConfigFile, config.Name, false)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if inheritedConfigSetting != nil {
			err = s.addSwarmConfigToService(ctx, app.ServiceID, inheritedConfigSetting.MustAsConfigFile().SwarmRef)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the config
	err = s.dockerManager.ConfigRemove(ctx, config.SwarmRef.ConfigID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	config.SwarmRef.ConfigID = ""
	config.SwarmRef.ConfigName = ""

	return nil
}

func (s *service) UpdateSwarmConfig(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldConfig, newConfig *entity.ConfigFile,
) (err error) {
	// Remove the old config from services then delete it from the swarm
	err = s.DeleteSwarmConfig(ctx, db, app, oldConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create a config in the swarm then add it to the services
	err = s.CreateSwarmConfig(ctx, db, app, newConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) addSwarmConfigToService(
	ctx context.Context,
	serviceID string,
	swarmRef *entity.SwarmConfigRef,
) (err error) {
	if serviceID == "" || swarmRef == nil || swarmRef.ConfigID == "" {
		return nil
	}

	swarmSvc, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Only add the config to the swarm service when the target file name is not used
	// by another config.
	for _, sec := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs {
		if sec.File != nil && sec.File.Name == swarmRef.File.Name {
			return nil
		}
	}

	swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs = append(swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs,
		&swarm.ConfigReference{
			File: &swarm.ConfigReferenceFileTarget{
				Name: swarmRef.File.Name,
				UID:  swarmRef.File.UID,
				GID:  swarmRef.File.GID,
				Mode: swarmRef.File.Mode.ToFileMode(),
			},
			ConfigID:   swarmRef.ConfigID,
			ConfigName: swarmRef.ConfigName,
		})

	_, err = s.dockerManager.ServiceUpdate(ctx, serviceID, &swarmSvc.Version, &swarmSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
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

	swarmSvc, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	hasChanges := false
	updateConfigs := make([]*swarm.ConfigReference, 0, len(swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs))
	for _, sec := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs {
		if swarmRef.ConfigID == sec.ConfigID {
			hasChanges = true
			continue
		}
		updateConfigs = append(updateConfigs, sec)
	}
	if !hasChanges {
		return nil
	}

	swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs = updateConfigs
	_, err = s.dockerManager.ServiceUpdate(ctx, serviceID, &swarmSvc.Version, &swarmSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
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

	orphanSwarmSec, err := s.dockerManager.ConfigInspect(ctx, configNameOrID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.Wrap(err)
	}

	orphanSwarmRef := &entity.SwarmConfigRef{
		File:       &entity.SwarmRefFileTarget{},
		ConfigID:   orphanSwarmSec.ID,
		ConfigName: orphanSwarmSec.Spec.Name,
	}

	// Remove the config from the swarm service of the app
	err = s.removeSwarmConfigFromService(ctx, app.ServiceID, orphanSwarmRef)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if app.ParentID == "" {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.removeSwarmConfigFromService(ctx, childApp.ServiceID, orphanSwarmRef)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the config
	err = s.dockerManager.ConfigRemove(ctx, orphanSwarmSec.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
