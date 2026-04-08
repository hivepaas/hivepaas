package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) UpdateCronJobMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.UpdateCronJobMetaReq,
) (*cronjobdto.UpdateCronJobMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingMetaData,
			_ *settings.PersistingSettingMetaData,
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

	return &cronjobdto.UpdateCronJobMetaResp{}, nil
}
