package repocheckoutserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filearchiver"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	repoCacheMinUpdateInterval       = time.Hour * 24
	repoCacheArchiveFormat           = filearchiver.ArchiveFormatTarLz4
	repoCacheArchiveCompressionLevel = filearchiver.CompressionLevelDefault
	repoCheckoutDurConsideredLong    = 3 * time.Minute
)

func (s *service) loadRepoCache(
	ctx context.Context,
	data *repoCheckoutData,
) (err error) {
	if data.NoCache {
		return nil
	}

	defer func() {
		if err != nil || recover() != nil {
			data.RepoCacheLoaded = false
			if err = s.resetCheckoutDir(data); err != nil {
				err = apperrors.New(err)
			} else {
				err = nil
			}
		}
		if data.RepoCacheLoaded {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Repository cache found. Try to use the cache.",
				tasklog.TsNow))
		}
	}()

	// NOTE: must use a separate `db` to establish another transaction
	err = transaction.Execute(ctx, s.db, func(db database.Tx) error {
		repoID := data.RepoSource.RepoID
		file, err := s.fileRepo.GetByKey(ctx, db, repoID,
			bunex.SelectFor("SHARE OF file"),
			bunex.SelectWhere("file.type = ?", base.FileTypeRepoCache),
			bunex.SelectWhere("file.status = ?", base.FileStatusActive),
			bunex.SelectWhere("file.object_id = ?", data.Project.ID),
		)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(err)
		}
		if file == nil || file.StorageType != base.FileStorageLocal {
			return nil
		}
		data.RepoCacheFile = file

		rootDir := config.Current.AppPath
		filePath := filepath.Join(rootDir, file.Path, file.Name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil
		}

		errStr, err := filearchiver.Decompress(filePath, data.CheckoutDir, filearchiver.ArchiveFormatAuto)
		if err != nil {
			return apperrors.New(err)
		}
		s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

		data.RepoCacheLoaded = true
		return nil
	})
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) saveRepoCache(
	ctx context.Context,
	data *repoCheckoutData,
) (err error) {
	// Cache is not enabled
	if data.NoCache {
		return nil
	}

	timeNow := timeutil.NowUTC()

	// We don't want to update the cache file too frequently as that requires file compression
	// which is a time-consuming action.
	shouldCache := !data.RepoCacheLoaded ||
		data.RepoCacheFile == nil ||
		timeNow.Sub(data.RepoCacheFile.UpdatedAt) >= repoCacheMinUpdateInterval ||
		// When checkout time is long, that often means a large amount of data has been downloaded
		data.CheckoutDuration > repoCheckoutDurConsideredLong

	if !shouldCache {
		return nil
	}

	var newCacheFile *entity.File
	if data.RepoCacheFile != nil {
		newCacheFile = new(*data.RepoCacheFile)
	}
	if newCacheFile == nil {
		newCacheFile = &entity.File{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeProject,
			ObjectID:    data.Project.ID,
			Type:        base.FileTypeRepoCache,
			Status:      base.FileStatusActive,
			Key:         data.RepoSource.RepoID,
			Path:        config.Current.DataPathSystemCacheRepos().RelPath(),
			Mimetype:    "application/octet-stream",
			StorageType: base.FileStorageLocal,
		}
	}
	for {
		newCacheFile.Name = fmt.Sprintf("%v.%v%v", newCacheFile.ID, gofn.RandTokenAsHex(4), //nolint:mnd
			repoCacheArchiveFormat.FileExtDefault())
		if data.RepoCacheFile == nil || data.RepoCacheFile.Name != newCacheFile.Name {
			break
		}
	}
	newCacheFile.UpdatedAt = timeNow
	newCacheFile.Deleted = false

	rootDir := config.Current.AppPath
	newFilePath := filepath.Join(rootDir, newCacheFile.Path, newCacheFile.Name)
	fileEntitySaved := false

	defer func() {
		if err == nil && recover() == nil && fileEntitySaved {
			// Remove the old cache file as it becomes orphaned
			if data.RepoCacheFile != nil {
				oldFilePath := filepath.Join(rootDir, data.RepoCacheFile.Path, data.RepoCacheFile.Name)
				_ = os.RemoveAll(oldFilePath)
			}
		} else {
			// Remove the new cache file as saving file record in DB failed
			_ = os.RemoveAll(newFilePath)
		}
	}()

	errStr, err := filearchiver.Compress(data.CheckoutDir, newFilePath,
		repoCacheArchiveFormat, repoCacheArchiveCompressionLevel)
	if err != nil {
		return apperrors.New(err)
	}
	s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

	newCacheFileInfo, err := os.Stat(newFilePath)
	if err != nil {
		return apperrors.New(err)
	}
	newCacheFile.Size = newCacheFileInfo.Size()

	err = transaction.Execute(ctx, s.db, func(db database.Tx) error {
		repoID := data.RepoSource.RepoID
		file, err := s.fileRepo.GetByKey(ctx, db, repoID,
			bunex.SelectFor("UPDATE OF file"),
			bunex.SelectWhere("file.type = ?", base.FileTypeRepoCache),
			bunex.SelectWhere("file.status = ?", base.FileStatusActive),
			bunex.SelectWhere("file.object_id = ?", data.Project.ID),
		)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(err)
		}
		if file != nil && (file.ID != newCacheFile.ID || file.UpdateVer != newCacheFile.UpdateVer) {
			return nil
		}

		newCacheFile.UpdateVer++
		err = s.fileRepo.Upsert(ctx, db, newCacheFile,
			entity.FileUpsertingConflictCols, entity.FileUpsertingUpdateCols)
		if err != nil {
			return apperrors.New(err)
		}

		fileEntitySaved = true
		return nil
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
