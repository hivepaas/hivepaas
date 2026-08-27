package apikeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) UpdateAPIKeyStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.UpdateAPIKeyStatusReq,
) (*apikeydto.UpdateAPIKeyStatusResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apikeydto.UpdateAPIKeyStatusResp{}, nil
}
