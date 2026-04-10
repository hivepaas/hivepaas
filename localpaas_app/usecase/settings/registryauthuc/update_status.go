package registryauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) UpdateRegistryAuthStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.UpdateRegistryAuthStatusReq,
) (*registryauthdto.UpdateRegistryAuthStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &registryauthdto.UpdateRegistryAuthStatusResp{}, nil
}
