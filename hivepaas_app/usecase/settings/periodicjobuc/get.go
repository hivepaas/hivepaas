package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

func (uc *UC) GetPeriodicJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.GetPeriodicJobReq,
) (*periodicjobdto.GetPeriodicJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := periodicjobdto.TransformPeriodicJob(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &periodicjobdto.GetPeriodicJobResp{
		Data: respData,
	}, nil
}
