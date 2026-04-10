package healthcheckuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) UpdateHealthcheckStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.UpdateHealthcheckStatusReq,
) (*healthcheckdto.UpdateHealthcheckStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &healthcheckdto.UpdateHealthcheckStatusResp{}, nil
}
