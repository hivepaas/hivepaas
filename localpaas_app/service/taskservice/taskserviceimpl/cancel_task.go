package taskserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *service) CancelTask(
	ctx context.Context,
	db database.Tx,
	taskID string,
) error {
	task, err := s.taskRepo.GetByID(ctx, db, "", taskID,
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
		err = s.taskRepo.Update(ctx, db, task,
			bunex.UpdateColumns("status", "updated_at"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}

	// Task is in-progress, send `cancel` command to the task executor
	err = s.CancelInProgressTask(ctx, taskID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) CancelInProgressTask(
	ctx context.Context,
	taskID string,
) error {
	// Get task info stored in redis
	taskInfo, err := s.taskInfoRepo.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(apperrors.ErrUnavailable).
				WithMsgLog("task info not found, please try again later")
		}
		return apperrors.Wrap(err)
	}

	if taskInfo.ControlDisabled {
		return apperrors.New(apperrors.ErrActionNotAllowed).
			WithMsgLog("task controlling is disabled")
	}

	err = s.taskControlRepo.Push(ctx, taskID, &cacheentity.TaskControl{
		ID:  taskID,
		Cmd: base.TaskCommandCancel,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
