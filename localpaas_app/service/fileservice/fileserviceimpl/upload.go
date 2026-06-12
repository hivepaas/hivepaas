package fileserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
)

// Upload uploads one or multiple files at the same time
func (s *service) Upload(
	ctx context.Context,
	db database.IDB,
	req *fileservice.UploadReq,
) (*fileservice.UploadResp, error) {
	files := make([]*entity.File, 0, len(req.Items))
	timeNow := timeutil.NowUTC()

	var filePath string
	switch req.FileType { //nolint
	case base.FileTypeBuildSource:
		filePath = config.Current.DataPathFilesUpload().RelPath()
	default:
		// Do nothing
	}

	for _, item := range req.Items {
		fileName := gofn.LastOr(strings.Split(item.FilePath, "/"), "")
		file := &entity.File{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       req.Scope.ScopeType(),
			ObjectID:    req.Scope.MainObjectID(),
			Status:      base.FileStatusActive,
			Type:        req.FileType,
			Name:        fileName,
			Path:        filePath,
			Size:        item.FileSize,
			Mimetype:    mime.TypeByExtension(strings.ToLower(filepath.Ext(fileName))),
			StorageType: req.StorageType,
			StorageID:   req.StorageID,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		if file.Type == base.FileTypeBuildSource {
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
			resp, err := s.uploadItem(ctx, r)
			if err == nil {
				responses[r.index] = resp
			}
			return err
		}, requests...)
	if err := errors.Join(gofn.MapValues(errMap)...); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &fileservice.UploadResp{Files: files}
	if req.SaveToDB {
		if err := s.fileRepo.InsertMulti(ctx, db, files); err != nil {
			return resp, apperrors.Wrap(err)
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
	_ context.Context,
	req *uploadItemReq,
) (*uploadItemResp, error) {
	if req.file.StorageType == base.FileStorageLocal {
		return s.uploadItemToLocal(req)
	}
	return nil, apperrors.NewNotImplemented()
}

func (s *service) uploadItemToLocal(
	req *uploadItemReq,
) (*uploadItemResp, error) {
	filePath := filepath.Join(config.Current.AppPath, req.file.Path, req.file.Name)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	// TODO: use sync.Pool with io.CopyBuffer here

	_, err = io.Copy(file, req.FileData)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer data from reader to file: %w", err)
	}
	return &uploadItemResp{}, nil
}
