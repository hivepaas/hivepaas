package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: uc.scheduleWithStatus,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.UpdateSchedJobStatusResp{}, nil
}

// scheduleWithStatus takes off the job's coming runs and, the job active, makes
// them anew: as the job is now, its new status saved. The setting loaded before
// was turned on with a job off, its runs made only at the next scan.
func (uc *UC) scheduleWithStatus(
	ctx context.Context,
	db database.Tx,
	_ *settings.UpdateSettingStatusData,
	persisting *settings.PersistingSettingStatusData,
) error {
	return hperrors.Wrap(uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, persisting.Setting, true))
}
