package syscleanupserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/services/aws/s3"
)

func (s *service) sysCleanupBackups(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	if !data.SysCleanupSettings.BackupCleanup.Enabled {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.BackupCleanup.Error = err.Error()
		}
	}()

	var errs []error

	// Remove old backup files in local
	if data.CleanupBackupInLocal != syscleanupservice.CleanupFlagFalse {
		errs = append(errs, s.sysCleanupLocalBackupFiles(ctx, db, data))
	}

	// Remove old backup files in cloud
	if data.CleanupBackupInLocal != syscleanupservice.CleanupFlagFalse {
		errs = append(errs, s.sysCleanupCloudBackupFiles(ctx, db, data))
	}

	return errors.Join(errs...)
}

func (s *service) sysCleanupLocalBackupFiles(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	if data.SysCleanupSettings.BackupCleanup.LocalBackupRetention < 0 { // No cleanup
		return nil
	}

	timeNow := timeutil.NowUTC()
	retention := data.SysCleanupSettings.BackupCleanup.LocalBackupRetention.ToDuration()

	deletingFiles, _, err := s.fileRepo.List(ctx, db, nil,
		bunex.SelectWhere("file.type = ?", base.FileTypeSystemBackup),
		bunex.SelectWhere("file.storage_type = ?", base.FileStorageLocal),
		bunex.SelectWhere("file.created_at < ?", timeNow.Add(-retention)),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, file := range deletingFiles {
		file.DeletedAt = timeNow
		data.TaskOutput.BackupCleanup.LocalBackupsDeleted++
	}
	err = s.fileRepo.UpsertMulti(ctx, db, deletingFiles, entity.FileUpsertingConflictCols,
		[]string{"deleted_at"})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete real files in local
	rootDir := config.Current.AppPath
	for _, file := range deletingFiles {
		filePath := filepath.Join(file.Path, file.Name)
		filePathAbs := filepath.Join(rootDir, filePath)
		err := os.Remove(filePathAbs)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to remove outdated backup file: "+
				filePath+" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Outdated backup file removed: "+filePath,
				tasklog.TsNow))
		}
	}

	return nil
}

func (s *service) sysCleanupCloudBackupFiles(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	if data.SysCleanupSettings.BackupCleanup.CloudBackupRetention < 0 { // No cleanup
		return nil
	}

	timeNow := timeutil.NowUTC()
	retention := data.SysCleanupSettings.BackupCleanup.CloudBackupRetention.ToDuration()

	deletingFiles, _, err := s.fileRepo.List(ctx, db, nil,
		bunex.SelectWhere("file.type = ?", base.FileTypeSystemBackup),
		bunex.SelectWhere("file.status = ?", base.FileStatusActive),
		bunex.SelectWhere("file.storage_type = ?", base.FileStorageCloud),
		bunex.SelectWhere("file.created_at < ?", timeNow.Add(-retention)),
		bunex.SelectRelation("Storage"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, file := range deletingFiles {
		file.DeletedAt = timeNow
		data.TaskOutput.BackupCleanup.CloudBackupsDeleted++
	}
	err = s.fileRepo.UpsertMulti(ctx, db, deletingFiles, entity.SettingUpsertingConflictCols,
		[]string{"deleted_at"})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Delete real files in cloud
	mapDelFuncByStorage := map[string]func(*entity.File) error{}

	getDelFunc := func(file *entity.File) (func(*entity.File) error, error) {
		delFunc, exists := mapDelFuncByStorage[file.Storage.ID]
		if exists {
			return delFunc, nil
		}
		if file.Storage == nil {
			return nil, apperrors.NewNotFound(apperrors.Fmt("Storage setting '%v'", file.Storage.ID))
		}

		switch base.CloudStorageKind(file.Storage.Kind) { //nolint:gocritic
		case base.CloudStorageKindS3:
			s3Client, err := s3.NewClientFromSetting(ctx, file.Storage)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			delFunc = func(file *entity.File) error {
				return s3Client.DeleteObject(ctx, file.Bucket, filepath.Join(file.Path, file.Name))
			}
			mapDelFuncByStorage[file.StorageID] = delFunc
		}
		return delFunc, nil
	}

	for _, file := range deletingFiles {
		filePath := filepath.Join(file.Path, file.Name)

		delFunc, err := getDelFunc(file)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to remove backup file in cloud: "+
				filePath+" with creating client error: "+err.Error(), tasklog.TsNow))
		}

		err = delFunc(file)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to remove backup file in cloud: "+
				filePath+" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Outdated backup file removed from cloud: "+filePath,
				tasklog.TsNow))
		}
	}
	return nil
}
