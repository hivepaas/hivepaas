package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) CancelTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.CancelTaskReq,
) (_ *taskdto.CancelTaskResp, err error) {
	var canceled bool
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		canceled, err = uc.taskService.CancelTask(ctx, db, req.ID, nil)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.CancelTaskResp{
		Data: &taskdto.CancelTaskDataResp{Canceled: canceled},
	}, nil
}
