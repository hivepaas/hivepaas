package userserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func (s *service) PersistUserData(ctx context.Context, db database.IDB,
	persistingData *userservice.PersistingUserData) error {
	// Persists data
	// Users
	err := s.userRepo.UpsertMulti(ctx, db, persistingData.UpsertingUsers,
		entity.UserUpsertingConflictCols, entity.UserUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Binaries
	err = s.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Remove accesses
	err = s.permissionManager.RemoveACLPermissions(ctx, db, persistingData.DeletingAccesses)
	if err != nil {
		return apperrors.New(err)
	}

	// Project/App/... accesses
	err = s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.UpsertingAccesses)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
