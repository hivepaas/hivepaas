package cronjobuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

func (uc *UC) ExecuteCronJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.ExecuteCronJobReq,
) (*cronjobdto.ExecuteCronJobResp, error) {
	req.Type = currentSettingType
	setting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.cronJobService.CreateCronJobTask(setting, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.ExecuteCronJobResp{
		Data: &cronjobdto.ExecuteCronJobDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}
