package appserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) DeleteApp(ctx context.Context, db database.IDB, app *entity.App) error {
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
	err = s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, base.SubjectTypeApp, appIDs)
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
	_, err = s.dockerManager.ServiceRemove(ctx, app.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Remove app config from traefik
	err = s.traefikService.RemoveAppConfig(ctx, app, nil)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
