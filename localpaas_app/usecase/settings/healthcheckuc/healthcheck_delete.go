package healthcheckuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

func (uc *UC) DeleteHealthcheck(
	ctx context.Context,
	auth *basedto.Auth,
	req *healthcheckdto.DeleteHealthcheckReq,
) (*healthcheckdto.DeleteHealthcheckResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &healthcheckdto.DeleteHealthcheckResp{}, nil
}
