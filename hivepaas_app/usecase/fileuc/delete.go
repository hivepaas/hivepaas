package fileuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

func (uc *UC) DeleteFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DeleteFileReq,
) (*filedto.DeleteFileResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		opts := []bunex.SelectQueryOption{
			bunex.SelectFor("UPDATE OF file"),
			bunex.SelectRelation("Storage"),
		}
		if req.Scope != nil {
			opts = append(opts, bunex.SelectWhere("file.scope = ?", *req.Scope))
		}
		if req.ObjectID != "" {
			opts = append(opts, bunex.SelectWhere("file.object_id = ?", req.ObjectID))
		}
		if len(req.Types) > 0 {
			opts = append(opts, bunex.SelectWhereIn("file.type IN (?)", req.Types...))
		}

		file, err := uc.fileRepo.GetByID(ctx, db, req.ID, opts...)
		if err != nil {
			return apperrors.Wrap(err)
		}

		deletePhysicalFile := false
		if req.DeletePermanently || (req.DeletePermanentlyIfLocal && file.StorageType == base.FileStorageLocal) {
			deletePhysicalFile = true
			_, err := uc.fileService.DeleteFileData(ctx, &fileservice.DeleteDataReq{
				File:     file,
				RetryMax: 2, //nolint:mnd
			})
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

		file.DeletedAt = time.Now()
		file.Deleted = deletePhysicalFile
		err = uc.fileRepo.Update(ctx, db, file, bunex.UpdateColumns("deleted", "deleted_at"))
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.DeleteFileResp{}, nil
}
