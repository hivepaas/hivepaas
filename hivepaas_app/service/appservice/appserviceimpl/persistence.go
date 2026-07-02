package appserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

func (s *service) PersistAppData(ctx context.Context, db database.IDB,
	persistingData *appservice.PersistingAppData) error {
	// Deletes all current linked data if configured
	err := s.appTagRepo.DeleteAllByApps(ctx, db, persistingData.AppsToDeleteTags)
	if err != nil {
		return apperrors.New(err)
	}

	// Persists data
	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// ResLinks
	err = s.resLinkRepo.UpsertMulti(ctx, db, persistingData.UpsertingResLinks,
		entity.ResLinkUpsertingConflictCols, entity.ResLinkUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Apps
	err = s.appRepo.UpsertMulti(ctx, db, persistingData.UpsertingApps,
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Tags
	err = s.appTagRepo.UpsertMulti(ctx, db, persistingData.UpsertingTags,
		entity.AppTagUpsertingConflictCols, entity.AppTagUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Deployments
	err = s.deploymentRepo.UpsertMulti(ctx, db, persistingData.UpsertingDeployments,
		entity.DeploymentUpsertingConflictCols, entity.DeploymentUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Tasks
	err = s.taskRepo.UpsertMulti(ctx, db, persistingData.UpsertingTasks,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
