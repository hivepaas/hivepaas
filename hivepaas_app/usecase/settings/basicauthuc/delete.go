package basicauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) DeleteBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.DeleteBasicAuthReq,
) (*basicauthdto.DeleteBasicAuthResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &basicauthdto.DeleteBasicAuthResp{}, nil
}
