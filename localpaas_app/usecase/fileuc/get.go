package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
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
		return nil, apperrors.Wrap(err)
	}

	respData, err := filedto.TransformFile(file)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.GetFileResp{
		Data: respData,
	}, nil
}
