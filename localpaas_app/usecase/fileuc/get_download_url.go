package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
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
		return nil, apperrors.Wrap(err)
	}

	resp, err := uc.fileService.GetDownloadURL(ctx, uc.db, auth, &fileservice.GetDownloadURLReq{
		File:         file,
		RequireLogin: req.RequireLogin,
		Expiration:   req.Expiration.ToDuration(),
		CloudPresign: req.CloudPresign,
		ViewInline:   req.ViewInline,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.GetFileDownloadURLResp{
		Data: &filedto.FileDownloadURLDataResp{URL: resp.URL, Expiration: req.Expiration},
	}, nil
}
