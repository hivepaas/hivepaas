package registryauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) ListRegistryAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.ListRegistryAuthReq,
) (*registryauthdto.ListRegistryAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := registryauthdto.TransformRegistryAuths(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &registryauthdto.ListRegistryAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
