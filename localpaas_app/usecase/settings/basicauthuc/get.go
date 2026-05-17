package basicauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) GetBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.GetBasicAuthReq,
) (*basicauthdto.GetBasicAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsBasicAuth().MustDecrypt()
	}

	respData, err := basicauthdto.TransformBasicAuth(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &basicauthdto.GetBasicAuthResp{
		Data: respData,
	}, nil
}
