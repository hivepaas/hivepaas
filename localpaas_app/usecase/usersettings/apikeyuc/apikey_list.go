package apikeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

func (uc *UC) ListAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.ListAPIKeyReq,
) (*apikeydto.ListAPIKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := apikeydto.TransformAPIKeys(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apikeydto.ListAPIKeyResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
