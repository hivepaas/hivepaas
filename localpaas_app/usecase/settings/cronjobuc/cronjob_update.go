package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) UpdateCronJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.UpdateCronJobReq,
) (*cronjobdto.UpdateCronJobResp, error) {
	req.Type = currentSettingType
	newJob := req.ToEntity()
	scheduleChanges := false
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: req.Name,
		AfterLoading: func(ctx context.Context, db database.Tx, data *settings.UpdateSettingData) error {
			job, err := data.Setting.AsCronJob()
			if err != nil {
				return apperrors.Wrap(err)
			}
			scheduleChanges = job.Schedule.Changed(newJob.Schedule)
			return nil
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(newJob.CronType)
			newJob.Schedule.OnChange(scheduleChanges) // call this to handle if the schedule changes
			err := pData.Setting.SetData(newJob)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := uc.taskQueue.ScheduleTasksForCronJob(ctx, db, data.Setting, scheduleChanges)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.UpdateCronJobResp{}, nil
}
