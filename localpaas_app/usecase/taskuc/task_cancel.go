package taskuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) CancelTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.CancelTaskReq,
) (*taskdto.CancelTaskResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		task, err := uc.taskRepo.GetByID(ctx, db, "", req.ID,
			bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
		)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}

		if task != nil {
			if !task.CanCancel() {
				return apperrors.New(apperrors.ErrStatusNotAllowAction)
			}
			task.Status = base.TaskStatusCanceled
			task.UpdatedAt = timeutil.NowUTC()
			err = uc.taskRepo.Update(ctx, db, task,
				bunex.UpdateColumns("status", "updated_at"),
			)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		}

		// Task is in-progress, send `cancel` command to the task executor
		_, err = uc.taskInfoRepo.Get(ctx, req.ID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return apperrors.New(apperrors.ErrUnavailable).
					WithMsgLog("task info not found, please try again later")
			}
			return apperrors.Wrap(err)
		}

		err = uc.taskControlRepo.Push(ctx, req.ID, &cacheentity.TaskControl{
			ID:  req.ID,
			Cmd: base.TaskCommandCancel,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.CancelTaskResp{}, nil
}
