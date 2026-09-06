package accesstokenuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) GetAccessToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.GetAccessTokenReq,
) (*accesstokendto.GetAccessTokenResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := accesstokendto.TransformAccessToken(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &accesstokendto.GetAccessTokenResp{
		Data: respData,
	}, nil
}
