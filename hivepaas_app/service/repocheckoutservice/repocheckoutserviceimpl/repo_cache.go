package repocheckoutserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filearchiver"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
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

	var cacheErr error
	defer func() {
		if err != nil || recover() != nil {
			cacheErr = err
			data.RepoCacheLoaded = false
			if err = s.resetCheckoutDir(data); err != nil {
				err = hperrors.Wrap(err)
			} else {
				err = nil
			}
		}
		if cacheErr != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Repository cache not used: "+cacheErr.Error(),
				tasklog.TsNow))
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
			bunex.SelectWhere("file.type = ?", base.FileTypeCache),
			bunex.SelectWhere("file.kind = ?", base.FileKindSourceCode),
			bunex.SelectWhere("file.status = ?", base.FileStatusActive),
			bunex.SelectWhere("file.object_id = ?", data.App.ProjectID),
		)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if file == nil || file.StorageType != base.FileStorageVolume {
			return nil
		}
		data.RepoCacheFile = file

		filePath, err := s.fetchCacheArchive(ctx, db, file, data.TempDir)
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil
		}
		if err != nil {
			return hperrors.Wrap(err)
		}
		defer os.Remove(filePath)

		errStr, err := filearchiver.Decompress(filePath, data.CheckoutDir, filearchiver.ArchiveFormatAuto)
		if err != nil {
			return hperrors.Wrap(err)
		}
		s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

		data.RepoCacheLoaded = true
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
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
			ObjectID:    data.App.ProjectID,
			Type:        base.FileTypeCache,
			Kind:        string(base.FileKindSourceCode),
			Status:      base.FileStatusActive,
			Key:         data.RepoSource.RepoID,
			Mimetype:    "application/octet-stream",
			StorageType: base.FileStorageVolume,
		}
	}
	for {
		newCacheFile.Name = fmt.Sprintf("%v.%v%v", newCacheFile.ID, gofn.RandTokenAsHex(4), //nolint:mnd
			repoCacheArchiveFormat.FileExtDefault())
		if data.RepoCacheFile == nil || data.RepoCacheFile.Name != newCacheFile.Name {
			break
		}
	}
	if err = s.placeRepoCache(ctx, data.App.ProjectID, newCacheFile); err != nil {
		return hperrors.Wrap(err)
	}
	newCacheFile.UpdatedAt = timeNow
	newCacheFile.Deleted = false

	archivePath := filepath.Join(data.TempDir, newCacheFile.Name)
	defer os.Remove(archivePath)
	fileStored, fileEntitySaved := false, false

	defer func() {
		cleanupCtx := context.WithoutCancel(ctx)
		if err == nil && recover() == nil && fileEntitySaved {
			// Remove the old cache file as it becomes orphaned
			if data.RepoCacheFile != nil {
				_ = s.fileService.Remove(cleanupCtx, s.db, data.RepoCacheFile)
			}
		} else if fileStored {
			// Remove the new cache file as saving file record in DB failed
			_ = s.fileService.Remove(cleanupCtx, s.db, newCacheFile)
		}
	}()

	errStr, err := filearchiver.Compress(data.CheckoutDir, archivePath,
		repoCacheArchiveFormat, repoCacheArchiveCompressionLevel)
	if err != nil {
		return hperrors.Wrap(err)
	}
	s.addCmdOutToLogs(ctx, errStr, err != nil, data.LogStore)

	newCacheFile.Size, err = s.storeCacheArchive(ctx, archivePath, newCacheFile)
	if err != nil {
		return hperrors.Wrap(err)
	}
	fileStored = true

	err = transaction.Execute(ctx, s.db, func(db database.Tx) error {
		repoID := data.RepoSource.RepoID
		file, err := s.fileRepo.GetByKey(ctx, db, repoID,
			bunex.SelectFor("UPDATE OF file"),
			bunex.SelectWhere("file.type = ?", base.FileTypeCache),
			bunex.SelectWhere("file.kind = ?", base.FileKindSourceCode),
			bunex.SelectWhere("file.status = ?", base.FileStatusActive),
			bunex.SelectWhere("file.object_id = ?", data.App.ProjectID),
		)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if file != nil && (file.ID != newCacheFile.ID || file.UpdateVer != newCacheFile.UpdateVer) {
			return nil
		}

		newCacheFile.UpdateVer++
		err = s.fileRepo.Upsert(ctx, db, newCacheFile,
			entity.FileUpsertingConflictCols, entity.FileUpsertingUpdateCols)
		if err != nil {
			return hperrors.Wrap(err)
		}

		fileEntitySaved = true
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// placeRepoCache puts a project's repository cache on the project's default
// volume, in HivePaaS's own directory.
func (s *service) placeRepoCache(ctx context.Context, projectID string, file *entity.File) error {
	volume, err := s.fileService.ProjectVolume(ctx, s.db, projectID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	file.StorageType = base.FileStorageVolume
	file.StorageID = volume.ID
	file.Path = fileservice.ProjectFilePath(fileservice.FilesDirRepoCache, file.Name)
	return nil
}

// storeCacheArchive writes the archive made in this process's temporary
// directory to the cache file's place, and gives its size.
func (s *service) storeCacheArchive(ctx context.Context, archivePath string, file *entity.File) (int64, error) {
	archive, err := os.Open(archivePath) //nolint:gosec // a file this process made
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	defer archive.Close()

	w, err := s.fileService.Create(ctx, s.db, file)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	size, err := io.Copy(w, archive)
	if err != nil {
		w.Abort(err)
		return 0, hperrors.Wrap(err)
	}
	if err = w.Close(); err != nil {
		return 0, hperrors.Wrap(err)
	}
	return size, nil
}

// fetchCacheArchive copies a cache file into tempDir, where the archiver can
// read it, and gives the copy's path.
func (s *service) fetchCacheArchive(
	ctx context.Context,
	db database.IDB,
	file *entity.File,
	tempDir string,
) (_ string, err error) {
	reader, err := s.fileService.Open(ctx, db, file)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	defer reader.Close()

	local, err := os.CreateTemp(tempDir, "repo-cache-*"+filepath.Ext(file.Name))
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	defer func() {
		_ = local.Close()
		if err != nil {
			_ = os.Remove(local.Name())
		}
	}()
	if _, err = io.Copy(local, reader); err != nil {
		return "", hperrors.Wrap(err)
	}
	return local.Name(), nil
}
