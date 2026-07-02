package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
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
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &taskdto.CancelTaskResp{
		Data: &taskdto.CancelTaskDataResp{Canceled: canceled},
	}, nil
}
