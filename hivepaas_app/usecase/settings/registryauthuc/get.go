package registryauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) GetRegistryAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.GetRegistryAuthReq,
) (*registryauthdto.GetRegistryAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsRegistryAuth().Decrypt(); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	respData, err := registryauthdto.TransformRegistryAuth(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &registryauthdto.GetRegistryAuthResp{
		Data: respData,
	}, nil
}
