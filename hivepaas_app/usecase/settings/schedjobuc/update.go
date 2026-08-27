package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	var oldJob *entity.SchedJob
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: newJob.GetRefObjectIDs(),
		AfterLoading: func(ctx context.Context, db database.Tx, data *settings.UpdateSettingData) error {
			err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App)
			if err != nil {
				return hperrors.Wrap(err)
			}
			oldJob, err = data.Setting.AsSchedJob()
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			if err := uc.checkPermissionPipeToApp(ctx, db, auth, newJob); err != nil {
				return hperrors.Wrap(err)
			}

			// Calculate upserting script objects if the scripts are too long
			upsertingScripts := uc.calcUpsertingScriptSettings(pData.Setting, newJob, oldJob)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			pData.Setting.Kind = string(newJob.JobType)
			err := pData.Setting.SetData(newJob)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, data.Setting,
				!oldJob.Schedule.Equal(newJob.Schedule))
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.UpdateSchedJobResp{}, nil
}
