package dbserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (s *service) MigrateData(
	ctx context.Context,
	db database.IDB,
) error {
	err := transaction.Execute(ctx, db, func(db database.Tx) error {
		migration, err := s.dataMigrationRepo.GetLatest(ctx, db,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if migration == nil {
			return apperrors.Wrap(apperrors.ErrInternal).WithMsgLog("no data migration found")
		}
		if migration.ID >= base.CurrentVersion { // no need migration
			return nil
		}

		err = s.migrateSettings(ctx, db)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Migration finishes
		newMigration := &entity.DataMigration{
			ID: base.CurrentVersion,
		}
		err = s.dataMigrationRepo.Insert(ctx, db, newMigration)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) migrateSettings(
	ctx context.Context,
	db database.IDB,
) error {
	offset, limit := 0, 100 //nolint:mnd
	for {
		settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
			bunex.SelectFor("UPDATE"),
			bunex.SelectWithDeleted(),
			bunex.SelectOffset(offset),
			bunex.SelectLimit(limit),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(settings) == 0 {
			break
		}
		offset += limit

		updatedSettings := make([]*entity.Setting, 0, len(settings))
		for _, setting := range settings {
			hasChange, err := setting.Migrate()
			if err != nil {
				return apperrors.Wrap(err)
			}
			if hasChange {
				updatedSettings = append(updatedSettings, setting)
			}
		}

		err = s.settingRepo.UpsertMulti(ctx, db, updatedSettings,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
