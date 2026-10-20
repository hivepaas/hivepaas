package dbserviceimpl

import (
	"context"
	"regexp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
			return hperrors.Wrap(err)
		}
		if migration == nil {
			return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("no data migration found")
		}
		if !needsMigration(migration.ID) {
			return nil
		}

		err = s.migrateSettings(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Migration finishes
		newMigration := &entity.DataMigration{
			ID:        base.CurrentVersion,
			AppliedAt: timeutil.NowUTC(),
		}
		err = s.dataMigrationRepo.Insert(ctx, db, newMigration)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
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
			return hperrors.Wrap(err)
		}
		if len(settings) == 0 {
			break
		}
		offset += limit

		updatedSettings := make([]*entity.Setting, 0, len(settings))
		for _, setting := range settings {
			hasChange, err := setting.Migrate()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if hasChange {
				updatedSettings = append(updatedSettings, setting)
			}
		}

		err = s.settingRepo.UpsertMulti(ctx, db, updatedSettings,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

// versionCodePattern is a version code, which sorts as the versions do.
var versionCodePattern = regexp.MustCompile(`^v[0-9]{6}$`)

// appliedCode is the version code a recorded data migration stands for. The
// table was made with a row named v1-20251019, which is v000001.
func appliedCode(id string) string {
	if versionCodePattern.MatchString(id) {
		return id
	}
	return base.VersionCodeV1
}

// needsMigration says whether data last migrated to applied is behind this
// HivePaaS's.
func needsMigration(applied string) bool {
	return appliedCode(applied) < base.CurrentVersion
}
