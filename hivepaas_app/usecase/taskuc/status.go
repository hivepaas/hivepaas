package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) GetTaskStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskStatusReq,
) (*taskdto.GetTaskStatusResp, error) {
	getResp, err := uc.taskService.GetTask(ctx, uc.db, &taskservice.GetTaskReq{
		Scope: req.Scope,
		ID:    req.ID,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.GetTaskStatusResp{
		Data: taskdto.TransformTaskStatus(getResp.Task, getResp.TaskInfo),
	}, nil
}
