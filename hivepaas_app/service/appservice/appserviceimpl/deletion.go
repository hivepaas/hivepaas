package appserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

func (s *service) DeleteApp(ctx context.Context, db database.IDB, app *entity.App) error {
	// Delete all child apps and their resources
	if app.ParentID == "" {
		childApps, _, err := s.appRepo.List(ctx, db, "", nil,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, childApp := range childApps {
			if err := s.DeleteApp(ctx, db, childApp); err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	// Delete ref resources in DB
	appIDs := []string{app.ID}

	// ACL permissions having the app ID as subject ID
	err := s.permissionManager.RemoveACLPermissionsBySubjects(ctx, db, base.SubjectTypeApp, appIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// App tags
	err = s.appTagRepo.DeleteAllByApps(ctx, db, appIDs)
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

	// Remove service for the app in docker swarm
	err = s.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	// Remove app config from traefik
	err = s.traefikService.RemoveAppConfig(ctx, app, nil)
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
