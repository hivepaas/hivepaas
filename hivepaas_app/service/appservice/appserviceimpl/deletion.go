package appserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) DeleteApp(ctx context.Context, db database.IDB, app *entity.App) error {
	// Delete all child apps and their resources
	if !app.IsChildApp() {
		childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			if err := s.DeleteApp(ctx, db, childApp); err != nil {
				return apperrors.Wrap(err).WithMsgLog("failed to delete child app %s", childApp.ID)
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
		return apperrors.Wrap(err)
	}
	for _, childApp := range logicalChildApps {
		if err := s.DeleteApp(ctx, db, childApp); err != nil {
			return apperrors.Wrap(err).WithMsgLog("failed to delete logical child app %s", childApp.ID)
		}
	}

	// TODO (high): remove secrets, config in docker

	// Remove service for the app in docker swarm
	err = s.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete ref resources in DB
	appIDs := []string{app.ID}

	// ACL permissions related to the app
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// App tags
	err = s.tagRepo.DeleteAllByObjects(ctx, db, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// App files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeApp, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete tasks and deployments with SKIP LOCKED to avoid blocking when an app is deleted
	// while a deployment is in-progress.
	// Any locked tasks/deployments will be cleaned up later by the daily cleanup cron job.

	// Tasks (must delete tasks before deployments)
	err = s.taskRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Deployments
	err = s.deploymentRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Remove app config from traefik
	_, err = s.traefikService.RemoveAppConfig(ctx, db, &traefikservice.RemoveAppConfigReq{
		App: app,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	app.DeletedAt = time.Now()
	app.UpdateVer++
	err = s.appRepo.Update(ctx, db, app, bunex.UpdateColumns("deleted_at", "update_ver"))
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
