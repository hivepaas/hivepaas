package registryauthrenewaluc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc/registryauthrenewaldto"
)

func (uc *UC) GetRegistryAuthRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthrenewaldto.GetRegistryAuthRenewalReq,
) (*registryauthrenewaldto.GetRegistryAuthRenewalResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := registryauthrenewaldto.TransformRegistryAuthRenewal(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registryauthrenewaldto.GetRegistryAuthRenewalResp{
		Data: respData,
	}, nil
}
