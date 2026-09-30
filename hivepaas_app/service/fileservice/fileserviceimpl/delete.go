package fileserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/cloudstorageservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
)

func (s *service) DeleteFileData(
	ctx context.Context,
	req *fileservice.DeleteDataReq,
) (_ *fileservice.DeleteDataResp, err error) {
	if req.RetryDelay <= 0 {
		req.RetryDelay = 3 * time.Second //nolint:mnd
	}
	// TODO: create an async task for deleting the file later
	err = gofn.ExecRetryCtx(ctx, func() error {
		return s.Remove(ctx, req.DB, req.File)
	}, req.RetryMax, req.RetryDelay)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &fileservice.DeleteDataResp{}, nil
}

func (s *service) removeCloud(ctx context.Context, db database.IDB, file *entity.File) error {
	if file.Storage == nil {
		return hperrors.NewInactive("Storage setting")
	}
	switch base.CloudStorageKind(file.Storage.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := cloudstorageservice.NewS3Client(ctx, db, file.Storage, nil)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = s3Client.DeleteObject(ctx, file.Bucket, file.Path); err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		return nil
	default:
		return hperrors.NewUnsupported("Storage type")
	}
}
