package healthcheckuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) UpdateHealthcheckStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.UpdateHealthcheckStatusReq,
) (*healthcheckdto.UpdateHealthcheckStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &healthcheckdto.UpdateHealthcheckStatusResp{}, nil
}
