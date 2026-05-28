package queueimpl

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
)

func (q *taskQueue) doCreateTasksForJobs(
	ctx context.Context,
) error {
	var newTasks []*entity.Task
	err := transaction.Execute(ctx, q.db, func(db database.Tx) (err error) {
		newTasks, err = q.createTasksForJobs(ctx, db, nil, q.config.Tasks.Queue.TaskCreateInterval)
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

func (q *taskQueue) createTasksForJobs(
	ctx context.Context,
	db database.Tx,
	jobIDs []string,
	withinDuration time.Duration,
) ([]*entity.Task, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSchedJob),
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
		schedJob, err := jobSetting.AsSchedJob()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		nextRuns, err := schedJob.Schedule.CalcNextRunsInRange(timeNow, timeNow.Add(withinDuration))
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		var lastSchedTime time.Time
		for _, nextRunAt := range nextRuns {
			lastSchedTime = nextRunAt
			task, err := q.schedJobService.CreateSchedJobTask(jobSetting, nextRunAt, timeNow)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			allNewTasks = append(allNewTasks, task)
		}

		if schedJob.Schedule.AdjustInitialTime(lastSchedTime) {
			jobSetting.MustSetData(schedJob)
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
