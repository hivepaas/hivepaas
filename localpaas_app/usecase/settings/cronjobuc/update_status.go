package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) UpdateCronJobStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.UpdateCronJobStatusReq,
) (*cronjobdto.UpdateCronJobStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
			_ *settings.PersistingSettingStatusData,
		) error {
			err := uc.taskQueue.ScheduleTasksForCronJob(ctx, db, data.Setting, true)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.UpdateCronJobStatusResp{}, nil
}
