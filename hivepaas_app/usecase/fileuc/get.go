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
	opts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Storage"),
	}
	if req.ObjectID != "" {
		opts = append(opts, bunex.SelectWhere("file.object_id = ?", req.ObjectID))
	}
	if len(req.Types) > 0 {
		opts = append(opts, bunex.SelectWhereIn("file.type IN (?)", req.Types...))
	}
	if len(req.Kinds) > 0 {
		opts = append(opts, bunex.SelectWhereIn("file.kind IN (?)", req.Kinds...))
	}

	file, err := uc.fileRepo.GetByID(ctx, uc.db, req.ID, opts...)
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
