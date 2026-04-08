package healthcheckuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) ListHealthcheck(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.ListHealthcheckReq,
) (*healthcheckdto.ListHealthcheckResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := healthcheckdto.TransformHealthchecks(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &healthcheckdto.ListHealthcheckResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
