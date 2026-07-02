package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

func (uc *UC) GetFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.GetFileReq,
) (*filedto.GetFileResp, error) {
	file, err := uc.fileRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("Storage"),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := filedto.TransformFile(file)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &filedto.GetFileResp{
		Data: respData,
	}, nil
}
