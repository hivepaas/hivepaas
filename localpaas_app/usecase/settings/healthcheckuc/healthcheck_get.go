package healthcheckuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) GetHealthcheck(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.GetHealthcheckReq,
) (*healthcheckdto.GetHealthcheckResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := healthcheckdto.TransformHealthcheck(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &healthcheckdto.GetHealthcheckResp{
		Data: respData,
	}, nil
}
