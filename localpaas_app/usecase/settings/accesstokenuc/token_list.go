package accesstokenuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) ListAccessToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.ListAccessTokenReq,
) (*accesstokendto.ListAccessTokenResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := accesstokendto.TransformAccessTokens(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &accesstokendto.ListAccessTokenResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
