package schedjobuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) ExecuteSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ExecuteSchedJobReq,
) (*schedjobdto.ExecuteSchedJobResp, error) {
	req.Type = currentSettingType
	setting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.schedJobService.CreateSchedJobTask(setting, time.Time{}, timeutil.NowUTC())
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

	return &schedjobdto.ExecuteSchedJobResp{
		Data: &schedjobdto.ExecuteSchedJobDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}
