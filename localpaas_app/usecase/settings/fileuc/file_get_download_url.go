package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
)

func (uc *UC) GetFileDownloadURL(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.GetFileDownloadURLReq,
) (*filedto.GetFileDownloadURLResp, error) {
	req.Type = currentSettingType
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, req.Scope, req.Type, req.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := uc.FileService.GetDownloadURL(ctx, uc.DB, auth, &fileservice.GetDownloadURLReq{
		File:         setting,
		RequireLogin: req.RequireLogin,
		Expiration:   req.Expiration,
		CloudPresign: req.CloudPresign,
		ViewInline:   req.ViewInline,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.GetFileDownloadURLResp{
		Data: &filedto.FileDownloadURLDataResp{URL: resp.URL, Expiration: timeutil.Duration(req.Expiration)},
	}, nil
}
