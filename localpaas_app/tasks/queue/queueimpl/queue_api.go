package queueimpl

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (q *taskQueue) ScheduleTask(
	ctx context.Context,
	tasks ...*entity.Task,
) error {
	if q.client == nil && q.server == nil {
		return apperrors.New(apperrors.ErrInternalServer).WithMsgLog("task queue is not initialized")
	}

	schedTasks := make([]*entity.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == base.TaskStatusDone || task.Status == base.TaskStatusCanceled {
			continue
		}
		schedTasks = append(schedTasks, task)
	}
	if len(schedTasks) == 0 {
		return nil
	}

	if q.client != nil { // Notify all workers to schedule the tasks
		if err := q.client.ScheduleTask(ctx, schedTasks...); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if q.server != nil { // Notify this worker to schedule the tasks
		if err := q.server.ScheduleTask(ctx, schedTasks...); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (q *taskQueue) UnscheduleTask(
	ctx context.Context,
	tasks ...*entity.Task,
) error {
	if len(tasks) == 0 {
		return nil
	}
	if q.client == nil && q.server == nil {
		return apperrors.New(apperrors.ErrInternalServer).WithMsgLog("task queue is not initialized")
	}

	taskIDs := entityutil.ExtractIDs(tasks)
	if q.client != nil { // Notify all workers to unschedule the tasks
		if err := q.client.UnscheduleTask(ctx, taskIDs...); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if q.server != nil { // Notify this worker to unschedule the tasks
		if err := q.server.UnscheduleTask(ctx, taskIDs...); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (q *taskQueue) ScheduleTasksForSchedJob(
	ctx context.Context,
	db database.Tx,
	jobSetting *entity.Setting,
	unscheduleCurrentTasks bool,
) error {
	return q.ScheduleTasksForSchedJobs(ctx, db, []*entity.Setting{jobSetting}, unscheduleCurrentTasks)
}

func (q *taskQueue) ScheduleTasksForSchedJobs(
	ctx context.Context,
	db database.Tx,
	jobSettings []*entity.Setting,
	unscheduleCurrentTasks bool,
) error {
	if unscheduleCurrentTasks {
		unschedulingTasks, err := q.loadCurrentTasksForUnscheduling(ctx, db, jobSettings)
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = q.taskRepo.UpsertMulti(ctx, db, unschedulingTasks,
			entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}
		// Unschedule the tasks from the queue, ignore error as tasks' status were updated in DB
		_ = q.UnscheduleTask(ctx, unschedulingTasks...)
	}

	activeJobIDs := make([]string, 0, len(jobSettings))
	for _, jobSetting := range jobSettings {
		if jobSetting.DeletedAt.IsZero() && jobSetting.IsActive() {
			activeJobIDs = append(activeJobIDs, jobSetting.ID)
		}
	}
	if len(activeJobIDs) == 0 {
		return nil
	}

	tasks, err := q.createTasksForJobs(ctx, db, activeJobIDs, q.config.Tasks.Queue.TaskCreateInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = q.ScheduleTask(ctx, tasks...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) loadCurrentTasksForUnscheduling(
	ctx context.Context,
	db database.IDB,
	jobs []*entity.Setting,
) ([]*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	tasks, _, err := q.taskRepo.List(ctx, db, "", nil,
		bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
		bunex.SelectWhereIn("task.target_id IN (?)", entityutil.ExtractIDs(jobs)...),
		bunex.SelectWhere("task.status != ?", base.TaskStatusDone),
		bunex.SelectWhere("task.run_at > ?", timeNow.Add(-10*24*time.Hour)), //nolint scan from 10 days ago
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	unschedulingTasks := make([]*entity.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.CanCancel() {
			task.Status = base.TaskStatusCanceled
			unschedulingTasks = append(unschedulingTasks, task)
			continue
		}
	}

	return unschedulingTasks, nil
}
