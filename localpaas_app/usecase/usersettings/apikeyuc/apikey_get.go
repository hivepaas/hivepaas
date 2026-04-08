package apikeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) GetAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.GetAPIKeyReq,
) (*apikeydto.GetAPIKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := apikeydto.TransformAPIKey(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apikeydto.GetAPIKeyResp{
		Data: respData,
	}, nil
}
