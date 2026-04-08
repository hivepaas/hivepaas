package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) GetTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskReq,
) (*taskdto.GetTaskResp, error) {
	getResp, err := uc.taskService.GetTask(ctx, uc.db, &taskservice.GetTaskReq{
		ID: req.ID,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTask(getResp.Task, getResp.TaskInfo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.GetTaskResp{
		Data: resp,
	}, nil
}
