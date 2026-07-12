package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) GetTaskStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskStatusReq,
) (*taskdto.GetTaskStatusResp, error) {
	getResp, err := uc.taskService.GetTask(ctx, uc.db, &taskservice.GetTaskReq{
		ID: req.ID,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.GetTaskStatusResp{
		Data: taskdto.TransformTaskStatus(getResp.Task, getResp.TaskInfo),
	}, nil
}
