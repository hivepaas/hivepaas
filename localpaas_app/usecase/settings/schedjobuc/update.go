package schedjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) UpdateSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.UpdateSchedJobReq,
) (*schedjobdto.UpdateSchedJobResp, error) {
	req.Type = currentSettingType
	newJob := req.ToEntity()
	scheduleChanges := false
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: newJob.GetRefObjectIDs(),
		AfterLoading: func(ctx context.Context, db database.Tx, data *settings.UpdateSettingData) error {
			job, err := data.Setting.AsSchedJob()
			if err != nil {
				return apperrors.Wrap(err)
			}
			scheduleChanges = !job.Schedule.Equal(newJob.Schedule)
			return nil
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(newJob.JobType)
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
			err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, data.Setting, scheduleChanges)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.UpdateSchedJobResp{}, nil
}
