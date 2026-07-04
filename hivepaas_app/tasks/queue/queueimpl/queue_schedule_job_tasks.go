package queueimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (q *taskQueue) doCreateTasksForJobs(
	ctx context.Context,
) error {
	var newTasks []*entity.Task
	err := transaction.Execute(ctx, q.db, func(db database.Tx) (err error) {
		newTasks, err = q.createTasksForJobs(ctx, db, nil, q.config.Tasks.Queue.TaskCreateInterval)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	}, transaction.NoRetry())
	if err != nil {
		return apperrors.New(err)
	}

	// Ignore error as tasks were inserted into DB, the next scan will schedule them again
	_ = q.ScheduleTask(ctx, newTasks...)
	return nil
}

func (q *taskQueue) createTasksForJobs(
	ctx context.Context,
	db database.Tx,
	jobIDs []string, // NOTE: no ID sent will create tasks for all current jobs in DB
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
		return nil, apperrors.New(err)
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
			return nil, apperrors.New(err)
		}
		nextRuns, err := schedJob.Schedule.CalcNextRunsInRange(timeNow, timeNow.Add(withinDuration))
		if err != nil {
			return nil, apperrors.New(err)
		}

		var lastSchedTime time.Time
		for _, nextRunAt := range nextRuns {
			lastSchedTime = nextRunAt
			task, err := q.schedJobService.CreateSchedJobTask(jobSetting, nextRunAt, timeNow)
			if err != nil {
				return nil, apperrors.New(err)
			}
			allNewTasks = append(allNewTasks, task)
		}

		if schedJob.Schedule.SetLastSchedTime(lastSchedTime) {
			jobSetting.MustSetData(schedJob)
			updatingJobSettings = append(updatingJobSettings, jobSetting)
		}
	}

	for _, chunk := range gofn.Chunk(allNewTasks, 10000) { //nolint:mnd
		err = q.taskRepo.UpsertMulti(ctx, db, chunk, entity.TaskUpsertingConflictColsByUK, nil)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	err = q.settingRepo.UpsertMulti(ctx, db, updatingJobSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return allNewTasks, nil
}
