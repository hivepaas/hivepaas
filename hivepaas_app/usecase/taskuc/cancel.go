package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) CancelTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.CancelTaskReq,
) (_ *taskdto.CancelTaskResp, err error) {
	var canceled bool
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		// The scope goes in, so the task has to be one this caller could have
		// reached: the route authorizes against a project, an environment or an
		// app, and without this the id alone decided what got canceled.
		task, done, e := uc.taskService.CancelTask(ctx, db, req.Scope, req.ID, nil)
		if e != nil {
			return hperrors.Wrap(e)
		}
		canceled = done

		// After the cancel, so the entry can say which of the two outcomes
		// happened, and inside its transaction, so a record that cannot be
		// written takes the cancel down with it.
		//
		// One case outruns that rollback: a task already in flight is stopped by
		// a message pushed to redis, and no rollback reaches it. Undoing the push
		// is not on offer - the executor may have acted on it already - so what
		// is left is that the window is the width of one insert into a database
		// the same transaction is already writing to.
		return uc.recordTaskCancel(ctx, db, auth, task, canceled)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.CancelTaskResp{
		Data: &taskdto.CancelTaskDataResp{Canceled: canceled},
	}, nil
}
