package appserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) DeleteApp(ctx context.Context, db database.IDB, app *entity.App, removeStorage bool) error {
	// Delete all child apps and their resources
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			if err := s.DeleteApp(ctx, db, childApp, removeStorage); err != nil {
				return hperrors.Wrap(err).WithMsgLog("failed to delete child app %s", childApp.ID)
			}
		}
	}

	// Query all logical-child-apps to delete (a preview app can have logical-child-apps linked via res_links)
	logicalChildApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectJoin("JOIN res_links AS res_link ON res_link.dst_id = app.id"),
		bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeApp),
		bunex.SelectWhere("res_link.src_id = ?", app.ID),
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeLogicalChildApp),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, childApp := range logicalChildApps {
		if err := s.DeleteApp(ctx, db, childApp, removeStorage); err != nil {
			return hperrors.Wrap(err).WithMsgLog("failed to delete logical child app %s", childApp.ID)
		}
	}

	if err := s.deleteAppInDocker(ctx, app, removeStorage); err != nil {
		return hperrors.Wrap(err)
	}

	// Delete ref resources in DB
	appIDs := []string{app.ID}

	// ACL permissions related to the app
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// App tags
	err = s.tagRepo.DeleteAllByObjects(ctx, db, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// App files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllByScope(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Delete tasks and deployments with SKIP LOCKED to avoid blocking when an app is deleted
	// while a deployment is in-progress.
	// Any locked tasks/deployments will be cleaned up later by the daily cleanup cron job.

	// Tasks (must delete tasks before deployments)
	err = s.taskRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Deployments
	err = s.deploymentRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Remove app config from traefik
	_, err = s.traefikService.RemoveAppConfig(ctx, db, &traefikservice.RemoveAppConfigReq{
		App: app,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	app.DeletedAt = time.Now()
	app.UpdateVer++
	err = s.appRepo.Update(ctx, db, app, bunex.UpdateColumns("deleted_at", "update_ver"))
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) getDockerSecretsAndConfigs(
	ctx context.Context,
	app *entity.App,
	service *swarm.Service, // can be nil
) ([]*swarm.SecretReference, []*swarm.ConfigReference, error) {
	if service == nil {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				return nil, nil, nil
			}
			return nil, nil, hperrors.Wrap(err)
		}
		service = &inspect.Service
	}

	if service.Spec.TaskTemplate.ContainerSpec == nil {
		return nil, nil, nil
	}
	secrets := service.Spec.TaskTemplate.ContainerSpec.Secrets
	configs := service.Spec.TaskTemplate.ContainerSpec.Configs

	return secrets, configs, nil
}

// deleteAppInDocker removes everything the app has outside the database: its
// service, the secrets and config files created for it, and - when asked - the
// directories it kept its data in.
//
// What has to be read has to be read before the service goes: the secrets, the
// configs and the mounts are all recorded on it, and afterwards nothing says
// where they were.
func (s *service) deleteAppInDocker(ctx context.Context, app *entity.App, removeStorage bool) error {
	if app.ServiceID == "" {
		return nil
	}
	secrets, configs, err := s.getDockerSecretsAndConfigs(ctx, app, nil)
	if err != nil {
		return hperrors.Wrap(err)
	}
	var mounts []mount.Mount
	if removeStorage {
		if mounts, err = s.getAppMounts(ctx, app); err != nil {
			return hperrors.Wrap(err)
		}
	}

	err = s.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Now that nothing is holding them open. Neither of these is fatal: what they
	// leave behind is something to clean up by hand, and failing here instead
	// would leave an app half deleted.
	_ = s.deleteDockerSecretsAndConfigs(ctx, secrets, configs)
	if removeStorage {
		_ = s.volumeService.RemoveAppStorage(ctx, mounts)
	}
	return nil
}

// getAppMounts reads the volumes an app has mounted, from the service itself: it
// is what the app actually ran with, which the settings need not still agree
// with.
func (s *service) getAppMounts(ctx context.Context, app *entity.App) ([]mount.Mount, error) {
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	if inspect.Service.Spec.TaskTemplate.ContainerSpec == nil {
		return nil, nil
	}
	return inspect.Service.Spec.TaskTemplate.ContainerSpec.Mounts, nil
}

func (s *service) deleteDockerSecretsAndConfigs(
	ctx context.Context,
	secrets []*swarm.SecretReference,
	configs []*swarm.ConfigReference,
) error {
	configIDs := make([]string, 0, len(configs))
	for _, config := range configs {
		configIDs = append(configIDs, config.ConfigID)
	}
	e1 := s.clusterSecretService.ConfigsRemove(ctx, configIDs, clusterservice.ItemRemovalRetryMax, 0)

	secretIDs := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		secretIDs = append(secretIDs, secret.SecretID)
	}
	e2 := s.clusterSecretService.SecretsRemove(ctx, secretIDs, clusterservice.ItemRemovalRetryMax, 0)

	err := errors.Join(e1, e2)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
