package cronjobuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) GetCronJobTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.GetCronJobTaskReq,
) (*cronjobdto.GetCronJobTaskResp, error) {
	req.Type = currentSettingType
	jobSetting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.JobID, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	getResp, err := uc.taskService.GetTask(ctx, uc.DB, &taskservice.GetTaskReq{
		ID:       req.TaskID,
		TargetID: jobSetting.ID,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTask(getResp.Task, getResp.TaskInfo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.GetCronJobTaskResp{
		Data: resp,
	}, nil
}
