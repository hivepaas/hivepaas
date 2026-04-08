package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

const (
	currentSettingType    = base.SettingTypeCronJob
	currentSettingVersion = entity.CurrentCronJobVersion
)

func (uc *UC) CreateCronJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.CreateCronJobReq,
) (*cronjobdto.CreateCronJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			cronJob := req.ToEntity()
			pData.Setting.Kind = string(cronJob.CronType)
			if err := pData.Setting.SetData(cronJob); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			err := uc.taskQueue.ScheduleTasksForCronJob(ctx, db, pData.Setting, false)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.CreateCronJobResp{
		Data: resp.Data,
	}, nil
}
