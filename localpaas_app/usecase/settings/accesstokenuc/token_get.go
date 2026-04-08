package accesstokenuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) GetAccessToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.GetAccessTokenReq,
) (*accesstokendto.GetAccessTokenResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsAccessToken().MustDecrypt()
	respData, err := accesstokendto.TransformAccessToken(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &accesstokendto.GetAccessTokenResp{
		Data: respData,
	}, nil
}
