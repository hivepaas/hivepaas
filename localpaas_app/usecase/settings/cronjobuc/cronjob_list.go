package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) ListCronJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.ListCronJobReq,
) (*cronjobdto.ListCronJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := cronjobdto.TransformCronJobs(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.ListCronJobResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
