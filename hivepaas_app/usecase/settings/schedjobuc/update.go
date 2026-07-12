package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
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
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, data.ScopeApp); err != nil {
				return apperrors.Wrap(err)
			}
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
			if err := uc.checkPermissionPipeToApp(ctx, db, auth, newJob); err != nil {
				return apperrors.Wrap(err)
			}
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
