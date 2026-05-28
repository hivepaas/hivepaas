package schedjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) GetSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.GetSchedJobReq,
) (*schedjobdto.GetSchedJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := schedjobdto.TransformSchedJob(resp.Data, resp.RefObjects, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.GetSchedJobResp{
		Data: respData,
	}, nil
}
