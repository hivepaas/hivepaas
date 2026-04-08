package oauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) ListOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.ListOAuthReq,
) (*oauthdto.ListOAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &oauthdto.OAuthTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := oauthdto.TransformOAuths(resp.Data, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &oauthdto.ListOAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
