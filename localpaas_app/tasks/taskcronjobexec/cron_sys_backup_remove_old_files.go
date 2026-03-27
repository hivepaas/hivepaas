package taskcronjobexec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (e *Executor) sysBackupRemoveOldFiles(
	ctx context.Context,
	db database.IDB,
	sysBackup *entity.SystemBackup,
	data *sysBackupTaskData,
) (err error) {
	allBackupFiles, err := os.ReadDir(data.BackupSaveDir)
	if err != nil {
		return apperrors.Wrap(err)
	}

	backupRetention := sysBackup.LocalBackupRetention.ToDuration()
	oldestTime := timeutil.NowUTC().Add(-backupRetention)
	backupFilenamesToDelete := make([]string, 0)
	upsertingFileSettings := make([]*entity.Setting, 0)

	for _, entry := range allBackupFiles {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if backupRetention == 0 {
			backupFilenamesToDelete = append(backupFilenamesToDelete, filename)
			continue
		}
		fileTime := sysBackupParseFileTime(filename)
		if fileTime.IsZero() || fileTime.Before(oldestTime) {
			err := os.Remove(filepath.Join(data.BackupSaveDir, filename))
			if err != nil {
				_ = data.LogStore.Add(ctx, applog.NewOutFrame("Failed to remove outdated backup file: "+
					filename+" with error: "+err.Error(), applog.TsNow))
			} else {
				_ = data.LogStore.Add(ctx, applog.NewOutFrame("Outdated backup file removed: "+filename,
					applog.TsNow))
				backupFilenamesToDelete = append(backupFilenamesToDelete, filename)
			}
		}
	}

	if len(backupFilenamesToDelete) > 0 {
		deletingFileSettings, _, err := e.settingRepo.List(ctx, db, base.NewSettingScopeGlobal(), nil,
			bunex.SelectWhere("setting.type = ?", base.SettingTypeFile),
			bunex.SelectWhere("setting.kind = ?", base.FileKindSystemBackup),
			bunex.SelectWhere("setting.data->>'storageType' = ?", base.FileStorageLocal),
			bunex.SelectWhereIn("setting.name IN (?)", backupFilenamesToDelete...),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, setting := range deletingFileSettings {
			setting.DeletedAt = data.TimeNow
			upsertingFileSettings = append(upsertingFileSettings, setting)
		}
	}

	if data.LocalOutFile != nil && backupRetention != 0 {
		upsertingFileSettings = append(upsertingFileSettings, data.LocalOutFile)
	}
	if data.RemoteOutFile != nil {
		upsertingFileSettings = append(upsertingFileSettings, data.RemoteOutFile)
	}

	err = e.settingRepo.UpsertMulti(ctx, db, upsertingFileSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func sysBackupParseFileTime(filename string) time.Time {
	filename = strings.TrimPrefix(filename, sysBackupFilePrefix)
	timeStr, _, _ := strings.Cut(filename, ".")
	dt, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return dt
}
