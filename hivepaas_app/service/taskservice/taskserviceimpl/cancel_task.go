package taskserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) CancelTask(
	ctx context.Context,
	db database.Tx,
	scope *entity.ObjectScope,
	taskID string,
	validatingTargetID *string,
) (_ *entity.Task, canceled bool, err error) {
	// Who the task belongs to is settled here, on a plain read, and not on the
	// locking read below.
	//
	// The locking one cannot answer it: SKIP LOCKED replies to a row a worker is
	// holding with nothing at all, which is indistinguishable from a row that was
	// never there - and a running task is exactly the row a cancel is usually
	// aimed at. Deciding on that answer sent every id nobody could prove anything
	// about down the in-progress path, where it was canceled without a check. An
	// id from another project reads as "busy" just as convincingly as one's own.
	//
	// A plain read does not block on the worker's lock, so the row comes back
	// either way, and the scope filter carries the same inheritance the task
	// listings use: a project's caller reaches the tasks of its environments and
	// apps, and nothing else.
	task, err := s.taskRepo.GetByID(ctx, db, scope, "", taskID)
	if err != nil {
		// Not found included, and it is the answer an id outside the scope gets:
		// what the caller may not reach should not be distinguishable from what
		// does not exist.
		return nil, false, hperrors.Wrap(err)
	}
	if validatingTargetID != nil && *validatingTargetID != task.TargetID {
		return nil, false, hperrors.NewNotFound("Task").WithMsgLog("unmatched task target id")
	}

	// Taken again, for update this time, because the row has to be held while it
	// is written. Missing here no longer means "who knows": it means a worker is
	// holding it, which is the in-progress case below.
	lockedTask, err := s.taskRepo.GetByID(ctx, db, nil, "", taskID,
		bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, false, hperrors.Wrap(err)
	}

	if lockedTask != nil {
		if !lockedTask.CanCancel() {
			return nil, false, hperrors.Wrap(hperrors.ErrActionNotAllowedByStatus)
		}
		lockedTask.Status = base.TaskStatusCanceled
		lockedTask.UpdatedAt = timeutil.NowUTC()
		err = s.taskRepo.Update(ctx, db, lockedTask,
			bunex.UpdateColumns("status", "updated_at"),
		)
		if err != nil {
			return nil, false, hperrors.Wrap(err)
		}
		// The row from the plain read, not the one just written: what a caller
		// reports is what was stopped, and the locked copy now says "canceled" in
		// place of the status that made this worth recording.
		return task, true, nil
	}

	// Task is in-progress, send `cancel` command to the task executor
	err = s.CancelInProgressTask(ctx, taskID)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	return task, false, nil
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
