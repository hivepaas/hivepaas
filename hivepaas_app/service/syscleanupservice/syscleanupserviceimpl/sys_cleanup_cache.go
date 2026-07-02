package syscleanupserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
)

const (
	repoCacheRetentionDefault = timeutil.Week
)

func (s *service) sysCleanupCache(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	if !data.SysCleanupSettings.CacheCleanup.Enabled {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.CacheCleanup.Error = err.Error()
		}
	}()

	var errs []error

	// Remove old repo cache files in local
	errs = append(errs, s.sysCleanupCacheRepoSource(ctx, db, data))

	return errors.Join(errs...)
}

func (s *service) sysCleanupCacheRepoSource(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	if data.CleanupCacheRepo == syscleanupservice.CleanupFlagFalse {
		return nil
	}
	timeNow := timeutil.NowUTC()
	retention := gofn.Coalesce(data.SysCleanupSettings.CacheCleanup.RepoCacheRetention.ToDuration(),
		repoCacheRetentionDefault)
	if data.CleanupCacheRepo == syscleanupservice.CleanupFlagForce {
		retention = 0
	}

	deletingFiles, _, err := s.fileRepo.List(ctx, db, nil,
		bunex.SelectWhere("file.type = ?", base.FileTypeRepoCache),
		bunex.SelectWhere("file.storage_type = ?", base.FileStorageLocal),
		bunex.SelectWhere("file.updated_at < ?", timeNow.Add(-retention)),
	)
	if err != nil {
		return apperrors.New(err)
	}

	for _, file := range deletingFiles {
		file.DeletedAt = timeNow
		data.TaskOutput.CacheCleanup.RepoCacheFilesDeleted++
		data.TaskOutput.CacheCleanup.RepoCacheSpaceReclaimed += uint64(file.Size) //nolint:gosec
	}
	err = s.fileRepo.UpsertMulti(ctx, db, deletingFiles, entity.FileUpsertingConflictCols,
		[]string{"deleted_at"}) //nolint:goconst
	if err != nil {
		return apperrors.New(err)
	}

	// Delete real files in local
	rootDir := config.Current.AppPath
	for _, file := range deletingFiles {
		filePath := filepath.Join(file.Path, file.Name)
		filePathAbs := filepath.Join(rootDir, filePath)
		err := os.Remove(filePathAbs)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Failed to remove outdated cache file: "+
				filePath+" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Outdated cache file removed: "+filePath,
				tasklog.TsNow))
		}
	}

	return nil
}
