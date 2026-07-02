package healthcheckuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) DeleteHealthcheck(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.DeleteHealthcheckReq,
) (*healthcheckdto.DeleteHealthcheckResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &healthcheckdto.DeleteHealthcheckResp{}, nil
}
