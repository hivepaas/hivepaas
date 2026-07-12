package basicauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) UpdateBasicAuthStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.UpdateBasicAuthStatusReq,
) (*basicauthdto.UpdateBasicAuthStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &basicauthdto.UpdateBasicAuthStatusResp{}, nil
}
