package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) GetSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.GetSchedJobReq,
) (*schedjobdto.GetSchedJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.IDB,
			data *settings.GetSettingData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := schedjobdto.TransformSchedJob(resp.Data, resp.RefObjects, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.GetSchedJobResp{
		Data: respData,
	}, nil
}
