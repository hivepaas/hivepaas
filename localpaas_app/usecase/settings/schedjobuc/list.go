package schedjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) ListSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ListSchedJobReq,
) (*schedjobdto.ListSchedJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := schedjobdto.TransformSchedJobs(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.ListSchedJobResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
