package appserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
)

func (s *service) DeleteApp(ctx context.Context, db database.IDB, app *entity.App) error {
	// Delete ref resources in DB
	appIDs := []string{app.ID}

	// ACL permissions having the app ID as subject ID
	err := s.permissionManager.RemoveACLPermissionsBySubjects(ctx, db, base.SubjectTypeApp, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// App tags
	err = s.appTagRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// App files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.ResourceTypeApp, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeApp, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Tasks (must delete tasks before deployments)
	err = s.taskRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Deployments
	err = s.deploymentRepo.DeleteAllByApps(ctx, db, appIDs)
	if err != nil {
		return apperrors.New(err)
	}

	// Remove service for the app in docker swarm
	err = s.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}

	// Remove app config from traefik
	err = s.traefikService.RemoveAppConfig(ctx, app, nil)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
