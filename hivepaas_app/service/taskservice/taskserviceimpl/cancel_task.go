package taskserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) CancelTask(
	ctx context.Context,
	db database.Tx,
	taskID string,
	validatingTargetID *string,
) (canceled bool, err error) {
	task, err := s.taskRepo.GetByID(ctx, db, "", taskID,
		bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return false, hperrors.Wrap(err)
	}

	if task != nil {
		if validatingTargetID != nil && *validatingTargetID != task.TargetID {
			return false, hperrors.NewNotFound("Task").WithMsgLog("unmatched task target id")
		}
		if !task.CanCancel() {
			return false, hperrors.Wrap(hperrors.ErrActionNotAllowedByStatus)
		}
		task.Status = base.TaskStatusCanceled
		task.UpdatedAt = timeutil.NowUTC()
		err = s.taskRepo.Update(ctx, db, task,
			bunex.UpdateColumns("status", "updated_at"),
		)
		if err != nil {
			return false, hperrors.Wrap(err)
		}
		return true, nil
	}

	// Task is in-progress, send `cancel` command to the task executor
	err = s.CancelInProgressTask(ctx, taskID)
	if err != nil {
		return false, hperrors.Wrap(err)
	}

	return false, nil
}

func (s *service) CancelInProgressTask(
	ctx context.Context,
	taskID string,
) error {
	// Get task info stored in redis
	taskInfo, err := s.taskInfoRepo.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(hperrors.ErrUnavailable).
				WithMsgLog("task info not found, please try again later")
		}
		return hperrors.Wrap(err)
	}

	if taskInfo.ControlDisabled {
		return hperrors.Wrap(hperrors.ErrActionNotAllowed).
			WithMsgLog("task controlling is disabled")
	}

	err = s.taskControlRepo.Push(ctx, taskID, &cacheentity.TaskControl{
		ID:  taskID,
		Cmd: base.TaskCommandCancel,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
