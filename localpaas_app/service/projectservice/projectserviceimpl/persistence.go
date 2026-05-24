package projectserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
)

func (s *service) PersistProjectData(ctx context.Context, db database.IDB,
	persistingData *projectservice.PersistingProjectData) error {
	// Deletes all current linked data if configured
	err := s.projectTagRepo.DeleteAllByProjects(ctx, db, persistingData.ProjectsToDeleteTags)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Persists data
	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// ACL Permissions
	err = s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.UpsertingACLPermissions)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Projects
	err = s.projectRepo.UpsertMulti(ctx, db, persistingData.UpsertingProjects,
		entity.ProjectUpsertingConflictCols, entity.ProjectUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Apps
	err = s.appRepo.UpsertMulti(ctx, db, persistingData.UpsertingApps,
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Project Tags
	err = s.projectTagRepo.UpsertMulti(ctx, db, persistingData.UpsertingTags,
		entity.ProjectTagUpsertingConflictCols, entity.ProjectTagUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
