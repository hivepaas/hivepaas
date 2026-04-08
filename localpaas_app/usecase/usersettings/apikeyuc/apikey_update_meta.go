package apikeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) UpdateAPIKeyMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.UpdateAPIKeyMetaReq,
) (*apikeydto.UpdateAPIKeyMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apikeydto.UpdateAPIKeyMetaResp{}, nil
}
