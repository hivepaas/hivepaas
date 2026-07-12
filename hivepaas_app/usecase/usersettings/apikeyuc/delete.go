package apikeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) DeleteAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.DeleteAPIKeyReq,
) (*apikeydto.DeleteAPIKeyResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.Wrap(apperrors.ErrUserDemoUnauthorized)
	}

	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		ExtraLoadOpts: []bunex.SelectQueryOption{
			bunex.SelectWhere("setting.object_id = ?", auth.User.ID),
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apikeydto.DeleteAPIKeyResp{}, nil
}
