package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) ListSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ListSchedJobReq,
) (*schedjobdto.ListSchedJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.IDB,
			data *settings.ListSettingData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, data.ScopeApp); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
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
