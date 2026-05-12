package queueimpl

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/gocronqueue"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redishelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

const (
	taskDefaultTimeout      = 3 * time.Hour
	taskInfoCacheExp        = 24 * time.Hour
	taskRetryMaxBackoff     = 24 * time.Hour
	taskControlCheckTimeout = 10 * time.Second
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
			return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}

		var execErr error
		defer func() {
			taskData.Done = true
			if err != nil {
				return
			}
			task.EndedAt = timeutil.NowUTC()
			if execErr != nil {
				task.Status = base.TaskStatusFailed
				if taskData.NonRetryable {
					task.Config.MaxRetry = task.Config.Retry
				}
				if task.CanRetry() {
					task.RetryAt = task.EndedAt.Add(calcExpBackoffRetry(task))
					rescheduleAt = &task.RetryAt
				} else {
					task.RetryAt = time.Time{}
				}
				_ = task.AddRun(&entity.TaskRun{
					StartedAt: task.StartedAt,
					EndedAt:   task.EndedAt,
					Error:     execErr.Error(),
				})
			} else {
				task.Status = gofn.If(taskData.Canceled, base.TaskStatusCanceled, base.TaskStatusDone)
			}
			// Post execution event
			if taskData.OnEndTransaction != nil {
				taskData.OnEndTransaction()
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
	if taskData != nil && taskData.OnPostTransaction != nil {
		taskData.OnPostTransaction()
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
		return nil, apperrors.Wrap(err)
	}

	// Task's target object must be active
	shouldCancelTask := false
	switch task.Type {
	case base.TaskTypeAppDeploy:
		shouldCancelTask = task.TargetDeployment == nil
	case base.TaskTypeCronJobExec:
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
		if taskData.Done || taskData.Canceled || ctx.Err() != nil {
			return
		}
		taskControl, err := redishelper.BLPopOne[*cacheentity.TaskControl](ctx, q.redisClient,
			key, taskControlCheckTimeout)
		if err != nil {
			continue
		}
		cmd := taskControl.Cmd
		if taskData.OnCommand != nil {
			taskData.OnCommand(cmd)
		}
		if !taskData.NonCancelable && cmd == base.TaskCommandCancel {
			taskData.Canceled = true
			return
		}
	}
}

func calcExpBackoffRetry(task *entity.Task) time.Duration {
	randDur := time.Duration(rand.Int31n(1000)) * time.Millisecond //nolint:mnd,gosec
	delay := task.Config.RetryDelay.ToDuration()
	if delay == 0 {
		return randDur
	}
	exp := 1.0
	if task.Config.Retry > 0 {
		exp = math.Pow(2, float64(task.Config.Retry)) //nolint:mnd
	}
	return min(time.Duration(exp*float64(delay))+randDur, taskRetryMaxBackoff)
}
