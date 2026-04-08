package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) DeleteCronJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.DeleteCronJobReq,
) (*cronjobdto.DeleteCronJobResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			_ *settings.PersistingSettingDeletionData,
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

	return &cronjobdto.DeleteCronJobResp{}, nil
}
