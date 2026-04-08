package oauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) GetOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.GetOAuthReq,
) (*oauthdto.GetOAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsOAuth().MustDecrypt()
	input := &oauthdto.OAuthTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := oauthdto.TransformOAuth(resp.Data, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &oauthdto.GetOAuthResp{
		Data: respData,
	}, nil
}
