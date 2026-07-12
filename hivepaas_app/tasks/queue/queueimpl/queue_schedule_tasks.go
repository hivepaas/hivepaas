package queueimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	missedTaskPeriod = 10 * time.Minute
)

func (q *taskQueue) findSchedulingTasks(
	ctx context.Context,
) ([]*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	scanFrom := timeNow.Add(-missedTaskPeriod)
	scanTo := timeNow.Add(q.config.Tasks.Queue.TaskCheckInterval)
	tasks, _, err := q.taskRepo.List(ctx, q.db, "", nil,
		bunex.SelectWhere("task.type != ?", base.TaskTypeHealthcheck),  // special tasks no need scheduling
		bunex.SelectWhere("task.type != ?", base.TaskTypeSystemUpdate), // special tasks
		bunex.SelectWhereGroup(
			// Not-started tasks
			bunex.SelectWhereGroup(
				bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
				bunex.SelectWhere("((task.run_at IS NULL AND task.created_at >= ?) "+
					"OR (task.run_at >= ? AND task.run_at < ?))", scanFrom, scanFrom, scanTo),
			),
			// Failed tasks need retry
			bunex.SelectWhereOrGroup(
				bunex.SelectWhere("task.status = ?", base.TaskStatusFailed),
				bunex.SelectWhere("task.retry_at IS NOT NULL"),
				bunex.SelectWhere("task.retry_at >= ?", scanFrom),
				bunex.SelectWhere("task.retry_at < ?", scanTo),
			),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	scheduleTasks := make([]*entity.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == base.TaskStatusFailed && !task.CanRetry() {
			continue
		}
		if task.IsNotStarted() && task.RunAt.Before(timeNow) && !q.shouldScheduleMissedTask(task, tasks, timeNow) {
			continue
		}
		scheduleTasks = append(scheduleTasks, task)
	}

	return scheduleTasks, nil
}

func (q *taskQueue) shouldScheduleMissedTask(
	missedTask *entity.Task,
	allTasks []*entity.Task,
	timeNow time.Time,
) bool {
	if missedTask.TargetID == "" { // This is a solo task
		return true
	}
	for _, task := range allTasks {
		if task.TargetID != missedTask.TargetID || task.ID == missedTask.ID {
			continue
		}
		runAt := task.ShouldRunAt()
		if runAt.IsZero() {
			runAt = timeNow
		}
		if task.IsNotStarted() && runAt.Before(timeNow) && runAt.After(missedTask.RunAt) {
			return false
		}
		// The next run is near, so ignore the missed task?
		if runAt.Sub(timeNow) < timeNow.Sub(missedTask.RunAt) {
			return false
		}
	}
	return true
}

func (q *taskQueue) canScheduleTask(task *entity.Task) bool {
	if task.Type == base.TaskTypeSystemUpdate { // System update task is run in the updater service
		return false
	}
	return true
}
