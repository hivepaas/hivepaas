package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/taskuc/taskdto"
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
