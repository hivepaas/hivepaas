package apikeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) DeleteAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.DeleteAPIKeyReq,
) (*apikeydto.DeleteAPIKeyResp, error) {
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
