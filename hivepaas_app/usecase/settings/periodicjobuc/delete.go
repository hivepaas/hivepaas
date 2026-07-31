package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

func (uc *UC) DeletePeriodicJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.DeletePeriodicJobReq,
) (*periodicjobdto.DeletePeriodicJobResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &periodicjobdto.DeletePeriodicJobResp{}, nil
}
