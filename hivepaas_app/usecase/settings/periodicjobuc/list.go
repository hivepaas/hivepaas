package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

func (uc *UC) ListPeriodicJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.ListPeriodicJobReq,
) (*periodicjobdto.ListPeriodicJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := periodicjobdto.TransformPeriodicJobs(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &periodicjobdto.ListPeriodicJobResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
