package queueimpl

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
)

const (
	missedTaskPeriod = 10 * time.Minute
)

func (q *taskQueue) doCreateTasks(
	ctx context.Context,
) error {
	var newTasks []*entity.Task
	err := transaction.Execute(ctx, q.db, func(db database.Tx) (err error) {
		newTasks, err = q.createTasks(ctx, db, nil, q.config.Tasks.Queue.TaskCreateInterval)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}, transaction.NoRetry())
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Ignore error as tasks were inserted into DB, the next scan will schedule them again
	_ = q.ScheduleTask(ctx, newTasks...)
	return nil
}

func (q *taskQueue) createTasks(
	ctx context.Context,
	db database.Tx,
	jobIDs []string,
	withinDuration time.Duration,
) ([]*entity.Task, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", base.SettingTypeCronJob),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectFor("UPDATE OF setting"),
	}
	if len(jobIDs) > 0 {
		opts = append(opts, bunex.SelectWhereIn("setting.id IN (?)", jobIDs...))
	}

	jobSettings, _, err := q.settingRepo.List(ctx, db, nil, nil, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(jobSettings) == 0 {
		return nil, nil
	}

	timeNow := timeutil.NowUTC()
	allNewTasks := make([]*entity.Task, 0, 20) //nolint:mnd
	updatingJobSettings := make([]*entity.Setting, 0, len(jobSettings))

	for _, jobSetting := range jobSettings {
		cronJob, err := jobSetting.AsCronJob()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		nextRuns, err := cronJob.Schedule.CalcNextRuns(timeNow, timeNow.Add(withinDuration), 0)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		var lastSchedTime time.Time
		for _, nextRunAt := range nextRuns {
			lastSchedTime = nextRunAt
			task, err := q.cronJobService.CreateCronJobTask(jobSetting, nextRunAt, timeNow)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			allNewTasks = append(allNewTasks, task)
		}

		if !lastSchedTime.Equal(cronJob.Schedule.LastSchedTime) {
			cronJob.Schedule.LastSchedTime = lastSchedTime
			jobSetting.MustSetData(cronJob)
			updatingJobSettings = append(updatingJobSettings, jobSetting)
		}
	}

	err = q.taskRepo.UpsertMulti(ctx, db, allNewTasks,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = q.settingRepo.UpsertMulti(ctx, db, updatingJobSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return allNewTasks, nil
}

func (q *taskQueue) findSchedulingTasks(
	ctx context.Context,
) ([]*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	scanFrom := timeNow.Add(-missedTaskPeriod)
	scanTo := timeNow.Add(q.config.Tasks.Queue.TaskCheckInterval)
	tasks, _, err := q.taskRepo.List(ctx, q.db, "", nil,
		bunex.SelectWhere("task.type != ?", base.TaskTypeHealthcheck), // special tasks no need scheduling
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
		if task.IsNotStarted() && task.RunAt.Before(timeNow) && !q.shouldRunMissedTask(task, tasks, timeNow) {
			continue
		}
		scheduleTasks = append(scheduleTasks, task)
	}

	return scheduleTasks, nil
}

func (q *taskQueue) shouldRunMissedTask(
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
	// System update task requires run mode `updater`
	if task.Type == base.TaskTypeSystemUpdate {
		return config.Current.RunMode == config.RunModeUpdater
	}
	return true
}
