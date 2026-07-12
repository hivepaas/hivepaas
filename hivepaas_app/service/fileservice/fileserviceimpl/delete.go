package fileserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

func (s *service) DeleteFileData(
	ctx context.Context,
	req *fileservice.DeleteDataReq,
) (_ *fileservice.DeleteDataResp, err error) {
	if req.RetryDelay <= 0 {
		req.RetryDelay = 3 * time.Second //nolint:mnd
	}
	switch req.File.StorageType {
	case base.FileStorageLocal:
		err = s.deleteLocalFile(ctx, req)
	case base.FileStorageCloud:
		err = s.deleteCloudFile(ctx, req)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &fileservice.DeleteDataResp{}, nil
}

func (s *service) deleteLocalFile(
	ctx context.Context,
	req *fileservice.DeleteDataReq,
) error {
	filePath := filepath.Join(config.Current.AppPath, req.File.Path)
	// TODO: create an async task for deleting the file later
	err := gofn.ExecRetryCtx(ctx, func() error {
		err := os.Remove(filePath)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}, req.RetryMax, req.RetryDelay)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) deleteCloudFile(
	ctx context.Context,
	req *fileservice.DeleteDataReq,
) (err error) {
	file := req.File
	if file.Storage == nil {
		return apperrors.NewInactive("Storage setting")
	}

	switch base.CloudStorageKind(file.Storage.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, file.Storage)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// TODO: create an async task for deleting the file later
		err = gofn.ExecRetryCtx(ctx, func() error {
			err = s3Client.DeleteObject(ctx, file.Bucket, file.Path)
			if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
				return apperrors.Wrap(err)
			}
			return nil
		}, req.RetryMax, req.RetryDelay)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil

	default:
		return apperrors.NewUnsupported("Storage type")
	}
}
