package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) ListTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.ListTaskReq,
) (*taskdto.ListTaskResp, error) {
	listResp, err := uc.taskService.ListTask(ctx, uc.db, &taskservice.ListTaskReq{
		TargetID: req.TargetID,
		Status:   req.Status,
		Search:   req.Search,
		Paging:   req.Paging,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTasks(listResp.Tasks, listResp.TaskInfoMap)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.ListTaskResp{
		Meta: &basedto.ListMeta{Page: listResp.PagingMeta},
		Data: resp,
	}, nil
}
