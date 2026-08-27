package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

func (uc *UC) UpdatePeriodicJobStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.UpdatePeriodicJobStatusReq,
) (*periodicjobdto.UpdatePeriodicJobStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &periodicjobdto.UpdatePeriodicJobStatusResp{}, nil
}
