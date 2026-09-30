package fileserviceimpl

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
)

// Upload uploads one or multiple files at the same time
func (s *service) Upload(
	ctx context.Context,
	db database.IDB,
	req *fileservice.UploadReq,
) (*fileservice.UploadResp, error) {
	files := make([]*entity.File, 0, len(req.Items))
	timeNow := timeutil.NowUTC()

	place, err := s.uploadPlace(ctx, db, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, item := range req.Items {
		fileName := gofn.LastOr(strings.Split(item.FilePath, "/"), "")
		mimetype := fileutil.TypeByExtension(filepath.Ext(fileName))
		fileID := gofn.Must(ulid.NewStringULID())
		file := &entity.File{
			ID:          fileID,
			Scope:       req.Scope.ScopeType,
			ObjectID:    req.Scope.ScopeObjectID(),
			Status:      base.FileStatusActive,
			Type:        req.FileType,
			Kind:        string(req.FileKind),
			Name:        fileName,
			Path:        place.path(fileID + "-" + fileName),
			Size:        item.FileSize,
			Mimetype:    gofn.Coalesce(mimetype, "application/octet-stream"),
			StorageType: req.StorageType,
			StorageID:   gofn.Coalesce(place.volumeID, req.StorageID),
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		if file.Type == base.FileTypeTmp {
			file.Name = file.ID + "-" + file.Name
		}
		files = append(files, file)
	}

	requests := make([]*uploadItemReq, 0, len(req.Items))
	responses := make([]*uploadItemResp, len(req.Items))
	for i, item := range req.Items {
		requests = append(requests, &uploadItemReq{
			UploadItemReq: item,
			index:         i,
			file:          files[i],
		})
	}

	errMap := gofn.ExecTaskFuncEx(ctx, req.ParallelUploads, true,
		func(ctx context.Context, r *uploadItemReq) error {
			resp, err := s.uploadItem(ctx, db, r)
			if err == nil {
				responses[r.index] = resp
			}
			return err
		}, requests...)
	if err := errors.Join(gofn.MapValues(errMap)...); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &fileservice.UploadResp{Files: files}
	if req.SaveToDB {
		if err := s.fileRepo.InsertMulti(ctx, db, files); err != nil {
			return resp, hperrors.Wrap(err)
		}
	}
	return resp, nil
}

type uploadItemReq struct {
	*fileservice.UploadItemReq
	index int
	file  *entity.File
}

type uploadItemResp struct {
}

func (s *service) uploadItem(
	ctx context.Context,
	db database.IDB,
	req *uploadItemReq,
) (*uploadItemResp, error) {
	if req.file.StorageType != base.FileStorageVolume {
		return nil, hperrors.NewNotImplemented()
	}
	w, err := s.Create(ctx, db, req.file)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if _, err = io.Copy(w, req.FileData); err != nil {
		w.Abort(err)
		return nil, hperrors.Wrap(err)
	}
	if err = w.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &uploadItemResp{}, nil
}

// uploadPlace is where an upload's files go: a file uploaded to an app, to its
// project's default volume, in HivePaaS's own directory; any other, to the
// system volume.
func (s *service) uploadPlace(ctx context.Context, db database.IDB, req *fileservice.UploadReq) (*uploadPlace, error) {
	if req.StorageType != base.FileStorageVolume {
		return &uploadPlace{path: func(name string) string { return name }}, nil
	}
	if req.Scope.ScopeType == base.ObjectScopeApp && req.App != nil {
		volume, err := s.ProjectVolume(ctx, db, req.App.ProjectID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		envKey := req.App.ProjectEnvID
		if req.App.ProjectEnv != nil {
			envKey = req.App.ProjectEnv.Key
		}
		return &uploadPlace{volumeID: volume.ID, path: func(name string) string {
			return fileservice.AppFilePath(fileservice.FilesDirUploads, envKey, req.App.Key, name)
		}}, nil
	}
	return &uploadPlace{path: func(name string) string {
		return filepath.Join(config.Current().DataPathFiles().RelPath(), name)
	}}, nil
}

type uploadPlace struct {
	volumeID string
	path     func(name string) string
}
