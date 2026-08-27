package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) DeleteSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.DeleteSchedJobReq,
) (*schedjobdto.DeleteSchedJobResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			currJob, err := data.Setting.AsSchedJob()
			if err != nil {
				return hperrors.Wrap(err)
			}
			// Calculate deleting script objects if there is any
			upsertingScripts := uc.calcUpsertingScriptSettings(data.Setting, nil, currJob)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, data.Setting, true)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.DeleteSchedJobResp{}, nil
}
