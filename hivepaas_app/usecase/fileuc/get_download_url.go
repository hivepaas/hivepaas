package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

func (uc *UC) GetFileDownloadURL(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.GetFileDownloadURLReq,
) (*filedto.GetFileDownloadURLResp, error) {
	file, err := uc.fileRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("Storage"),
		bunex.SelectWhere("file.status = ?", base.FileStatusActive),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := uc.fileService.GetDownloadURL(ctx, uc.db, auth, &fileservice.GetDownloadURLReq{
		File:         file,
		RequireLogin: req.RequireLogin,
		Expiration:   req.Expiration.ToDuration(),
		CloudPresign: req.CloudPresign,
		ViewInline:   req.ViewInline,
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &filedto.GetFileDownloadURLResp{
		Data: &filedto.FileDownloadURLDataResp{URL: resp.URL, Expiration: req.Expiration},
	}, nil
}
