package registryauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
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

	resp.Data.MustAsRegistryAuth().MustDecrypt()
	respData, err := registryauthdto.TransformRegistryAuth(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &registryauthdto.GetRegistryAuthResp{
		Data: respData,
	}, nil
}
