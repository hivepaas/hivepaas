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

func (s *service) RemoveConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	configs ...*entity.ConfigFile,
) (err error) {
	configRefs := make([]*entity.SwarmConfigRef, 0, len(configs))
	for _, config := range configs {
		if config == nil || config.SwarmRef == nil || config.SwarmRef.ConfigID == "" {
			continue
		}
		configRefs = append(configRefs, config.SwarmRef)
	}
	if len(configRefs) == 0 {
		return nil
	}

	// Remove the config items from the swarm service of the app
	err = s.removeSwarmConfigFromService(ctx, app.ServiceID, configRefs...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			err = s.removeSwarmConfigFromService(ctx, childApp.ServiceID, configRefs...)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Now delete the config items from docker
	for _, configRef := range configRefs {
		err = s.ConfigRemove(ctx, configRef.ConfigID, itemRemovalRetryMax, itemRemovalRetryDelay)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		configRef.ConfigID = ""
		configRef.ConfigName = ""
	}

	return nil
}

func (s *service) removeSwarmConfigFromService(
	ctx context.Context,
	serviceID string,
	swarmRefs ...*entity.SwarmConfigRef,
) (err error) {
	if serviceID == "" {
		return nil
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, serviceID, nil,
		func(_ int, swarmSvc *swarm.Service) (bool, error) {
			hasChanges := false
			updateConfigs := make([]*swarm.ConfigReference, 0, len(swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs))
			for _, swarmRef := range swarmRefs {
				if swarmRef == nil || swarmRef.ConfigID == "" {
					continue
				}
				for _, cfg := range swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs {
					if swarmRef.ConfigID == cfg.ConfigID {
						hasChanges = true
						continue
					}
					updateConfigs = append(updateConfigs, cfg)
				}
			}
			if !hasChanges {
				return false, nil
			}
			swarmSvc.Spec.TaskTemplate.ContainerSpec.Configs = updateConfigs
			return true, nil
		}, itemRemovalRetryMax, 0)
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

	inspect, err := s.dockerManager.ConfigInspect(ctx, configNameOrID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}

	// If this app is parent of some other apps, also remove the config from the child apps
	if !app.IsChildApp() {
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
	_, err = s.dockerManager.ConfigRemove(ctx, orphanSwarmConfig.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	return nil
}
