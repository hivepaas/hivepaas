package configfileuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

const (
	defaultDownloadTokenExp    = 30 * time.Second
	defaultDownloadTokenExpDev = 60 * time.Second
)

func (uc *UC) GetDownloadToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.GetDownloadTokenReq,
) (*configfiledto.GetDownloadTokenResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	expiration := req.Expiration.ToDuration()
	if expiration <= 0 {
		expiration = gofn.If(config.Current.IsDevEnv(), defaultDownloadTokenExpDev, defaultDownloadTokenExp)
	}
	token, err := uc.FileService.GenerateDownloadToken(auth.User.ID, resp.Data.ID, false, expiration)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &configfiledto.GetDownloadTokenResp{
		Data: &configfiledto.GetDownloadTokenDataResp{
			Token: token,
		},
	}, nil
}
