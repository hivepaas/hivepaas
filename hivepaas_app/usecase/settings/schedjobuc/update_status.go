package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) UpdateSchedJobStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.UpdateSchedJobStatusReq,
) (*schedjobdto.UpdateSchedJobStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, data.ScopeApp); err != nil {
				return apperrors.New(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
			_ *settings.PersistingSettingStatusData,
		) error {
			err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, data.Setting, true)
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &schedjobdto.UpdateSchedJobStatusResp{}, nil
}
