package oauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) GetOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.GetOAuthReq,
) (*oauthdto.GetOAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsOAuth().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	input := &oauthdto.OAuthTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := oauthdto.TransformOAuth(setting, input)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &oauthdto.GetOAuthResp{
		Data: respData,
	}, nil
}
