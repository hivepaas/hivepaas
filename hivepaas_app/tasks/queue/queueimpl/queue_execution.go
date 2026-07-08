package queueimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/gocronqueue"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

const (
	taskDefaultTimeout      = 3 * time.Hour
	taskInfoCacheExp        = 24 * time.Hour
	taskControlCheckTimeout = 7 * time.Second
)

func (q *taskQueue) RegisterExecutor(typ base.TaskType, execFunc queue.TaskExecFunc) {
	if !q.isWorkerMode() {
		return
	}
	if q.taskExecutorMap == nil {
		q.taskExecutorMap = make(map[base.TaskType]gocronqueue.TaskExecFunc, 5) //nolint:mnd
	}
	q.taskExecutorMap[typ] = func(taskID string, payload string) *time.Time {
		return q.executeTask(context.Background(), taskID, payload, execFunc)
	}
}

//nolint:gocognit
func (q *taskQueue) executeTask(
	ctx context.Context,
	taskID string,
	_ string,
	executorFunc func(context.Context, database.Tx, *queue.TaskExecData) error,
) (rescheduleAt *time.Time) {
	var taskData *queue.TaskExecData
	err := transaction.Execute(ctx, q.db, func(db database.Tx) (err error) {
		task, err := q.loadTask(ctx, db, taskID)
		if err != nil {
			return apperrors.New(err)
		}
		if task == nil {
			return nil
		}

		taskData = &queue.TaskExecData{
			Task: task,
		}
		taskTimeout := task.Config.Timeout.ToDuration()
		ctx, cancel := context.WithTimeout(ctx, gofn.Coalesce(taskTimeout, taskDefaultTimeout))
		defer cancel()
		taskData.CancelFunc = cancel

		// Put task to in-progress state
		task.StartedAt = timeutil.NowUTC()
		if task.RunAt.IsZero() {
			task.RunAt = task.StartedAt
		}
		if task.Status == base.TaskStatusFailed {
			task.Config.Retry++
		}

		// Check for control commands from users and other processing
		if !task.Config.ControlDisabled {
			go q.taskControlCheck(ctx, taskData)
		}

		// Mark the task as `in-progress` by inserting a new record in to redis
		taskInfo := &cacheentity.TaskInfo{
			ID:              task.ID,
			Status:          base.TaskStatusInProgress,
			ControlDisabled: task.Config.ControlDisabled,
			StartedAt:       task.StartedAt,
		}
		err = q.taskInfoRepo.Set(ctx, task.ID, taskInfo, taskInfoCacheExp)
		if err != nil {
			return apperrors.New(err)
		}

		var execErr error
		defer func() {
			taskData.TaskDone = true
			if err != nil {
				return
			}
			task.EndedAt = timeutil.NowUTC()
			if execErr != nil {
				task.Status = base.TaskStatusFailed
				if taskData.TaskNonRetryable {
					task.Config.MaxRetry = task.Config.Retry
				}
				if task.CanRetry() {
					task.RetryAt = task.EndedAt.Add(task.NextRetryDelay())
					rescheduleAt = &task.RetryAt
				} else {
					task.RetryAt = time.Time{}
				}
				_ = task.AddRun(&entity.TaskRun{
					StartedAt: task.StartedAt,
					EndedAt:   task.EndedAt,
					Error:     apperrors.GetErrorDetail(execErr, ""),
				})
			} else {
				task.Status = gofn.If(taskData.TaskCanceled, base.TaskStatusCanceled, base.TaskStatusDone)
			}
			// Post execution event
			if taskData.OnEndTransactionFunc != nil {
				taskData.OnEndTransactionFunc()
			}
			// Delete data in cache
			_ = q.taskInfoRepo.Del(ctx, task.ID)
			// Save tasks in DB
			err = q.taskRepo.Update(ctx, db, task)
		}()
		defer funcutil.EnsureNoPanic(&err) // Make sure we catch panic before the above defer

		execErr = executorFunc(ctx, db, taskData)
		return err //nolint:wrapcheck
	}, transaction.NoRetry())
	if err != nil {
		return rescheduleAt
	}

	// Post transaction event
	if taskData != nil && taskData.OnPostTransactionFunc != nil {
		taskData.OnPostTransactionFunc()
	}

	return rescheduleAt
}

func (q *taskQueue) loadTask(
	ctx context.Context,
	db database.IDB,
	taskID string,
) (*entity.Task, error) {
	task, err := q.taskRepo.GetByID(ctx, db, "", taskID,
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusFailed),
		bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
		bunex.SelectRelation("TargetJob"),
		bunex.SelectRelation("TargetDeployment"),
	)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) { // task not found, it's not error
			return nil, nil
		}
		return nil, apperrors.New(err)
	}

	// Task's target object must be active
	shouldCancelTask := false
	switch task.Type {
	case base.TaskTypeAppDeploy:
		shouldCancelTask = task.TargetDeployment == nil
	case base.TaskTypeSchedJobExec:
		shouldCancelTask = task.TargetJob == nil || !task.TargetJob.IsActive()
	case base.TaskTypeHealthcheck:
		// Do nothing
	case base.TaskTypeSystemUpdate:
		// Do nothing
	case base.TaskTypeDummy:
		// Do nothing
	}
	if shouldCancelTask {
		task.Status = base.TaskStatusCanceled
		task.UpdatedAt = timeutil.NowUTC()
		_ = q.taskRepo.Update(ctx, db, task, bunex.UpdateColumns("status", "updated_at"))
		return nil, nil
	}

	// Task not allow retrying
	if task.Status == base.TaskStatusFailed && !task.CanRetry() {
		return nil, nil
	}
	return task, nil
}

func (q *taskQueue) taskControlCheck(
	ctx context.Context,
	taskData *queue.TaskExecData,
) {
	key := fmt.Sprintf("task:%s:ctrl", taskData.Task.ID)
	defer func() {
		_ = recover()
		_ = redishelper.Del(ctx, q.redisClient, key)
	}()

	for {
		if taskData.TaskDone || taskData.TaskCanceled || ctx.Err() != nil {
			return
		}
		taskControl, err := redishelper.BLPopOne[*cacheentity.TaskControl](ctx, q.redisClient,
			key, taskControlCheckTimeout)
		if err != nil {
			continue
		}
		cmd := taskControl.Cmd
		if taskData.OnCommandFunc != nil {
			taskData.OnCommandFunc(cmd)
		}
		if !taskData.TaskNonCancelable && cmd == base.TaskCommandCancel {
			taskData.CancelFunc()
			taskData.TaskCanceled = true
			return
		}
	}
}
