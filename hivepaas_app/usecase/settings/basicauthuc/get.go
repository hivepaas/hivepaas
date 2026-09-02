package basicauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) GetBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.GetBasicAuthReq,
) (*basicauthdto.GetBasicAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsBasicAuth().Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	respData, err := basicauthdto.TransformBasicAuth(setting, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &basicauthdto.GetBasicAuthResp{
		Data: respData,
	}, nil
}
