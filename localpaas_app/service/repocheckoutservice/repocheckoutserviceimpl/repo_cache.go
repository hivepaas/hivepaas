package repocheckoutserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/filearchiver"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
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
				err = apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}
		if file == nil || file.StorageType != base.FileStorageLocal {
			return nil
		}
		data.RepoCache = file

		rootDir := config.Current.AppPath
		filePath := filepath.Join(rootDir, file.Path, file.Name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil
		}

		errStr, err := filearchiver.Decompress(filePath, data.CheckoutDir, filearchiver.ArchiveFormatAuto)
		if err != nil {
			return apperrors.Wrap(err)
		}
		s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

		data.RepoCacheLoaded = true
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
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
		data.RepoCache == nil ||
		timeNow.Sub(data.RepoCache.UpdatedAt) >= repoCacheMinUpdateInterval ||
		// When checkout time is long, that often means a large amount of data has been downloaded
		data.CheckoutDuration > repoCheckoutDurConsideredLong

	if !shouldCache {
		return nil
	}

	var newCacheFile *entity.File
	if data.RepoCache != nil {
		newCacheFile = new(*data.RepoCache)
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
		if data.RepoCache == nil || data.RepoCache.Name != newCacheFile.Name {
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
			if data.RepoCache != nil {
				oldFilePath := filepath.Join(rootDir, data.RepoCache.Path, data.RepoCache.Name)
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
		return apperrors.Wrap(err)
	}
	s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

	newCacheFileInfo, err := os.Stat(newFilePath)
	if err != nil {
		return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}
		if file != nil && (file.ID != newCacheFile.ID || file.UpdateVer != newCacheFile.UpdateVer) {
			return nil
		}

		newCacheFile.UpdateVer++
		err = s.fileRepo.Upsert(ctx, db, newCacheFile,
			entity.FileUpsertingConflictCols, entity.FileUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}

		fileEntitySaved = true
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
