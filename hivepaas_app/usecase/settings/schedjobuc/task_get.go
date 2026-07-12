package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) GetSchedJobTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.GetSchedJobTaskReq,
) (*schedjobdto.GetSchedJobTaskResp, error) {
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

	return &schedjobdto.GetSchedJobTaskResp{
		Data: resp,
	}, nil
}
