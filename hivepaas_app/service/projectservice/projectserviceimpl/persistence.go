package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

func (s *service) PersistProjectData(ctx context.Context, db database.IDB,
	persistingData *projectservice.PersistingProjectData) error {
	// Deletes all current linked data if configured
	err := s.tagRepo.DeleteAllByObjects(ctx, db, persistingData.ProjectsToDeleteTags)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Persists data
	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// ACL Permissions
	err = s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.UpsertingACLPermissions)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Projects
	err = s.projectRepo.UpsertMulti(ctx, db, persistingData.UpsertingProjects,
		entity.ProjectUpsertingConflictCols, entity.ProjectUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project envs
	err = s.projectEnvRepo.UpsertMulti(ctx, db, persistingData.UpsertingProjectEnvs,
		entity.ProjectEnvUpsertingConflictCols, entity.ProjectEnvUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Apps
	err = s.appRepo.UpsertMulti(ctx, db, persistingData.UpsertingApps,
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project Tags
	err = s.tagRepo.UpsertMulti(ctx, db, persistingData.UpsertingTags,
		entity.TagUpsertingConflictCols, entity.TagUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
