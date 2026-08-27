package apikeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		ExtraLoadOpts: []bunex.SelectQueryOption{
			bunex.SelectWhere("setting.object_id = ?", auth.User.ID),
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apikeydto.DeleteAPIKeyResp{}, nil
}
