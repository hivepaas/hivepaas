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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

const (
	taskHealthcheckLockKey      = "task:healthcheck:%v:lock"
	cacheHealthcheckSettingsExp = 5 * time.Minute
)

func (q *taskQueue) RegisterHealthcheckExecutor(execFunc queue.HealthcheckExecFunc) {
	if !q.isWorkerMode() {
		return
	}
	q.healthcheckExecutor = execFunc
}

func (q *taskQueue) doHealthcheck(
	ctx context.Context,
) error {
	if q.healthcheckExecutor == nil {
		return apperrors.NewUnavailable("Task executor function for healthcheck")
	}

	baseData := &queue.HealthcheckExecData{}
	jobSettings, err := q.loadHealthcheckData(ctx, q.db, baseData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if len(jobSettings) == 0 {
		return nil
	}

	timeNow := timeutil.NowUTC()
	savingTasks := make([]*entity.Task, 0, len(jobSettings))
	execFuncs := make([]func(ctx context.Context) error, 0, len(jobSettings))

	for _, jobSetting := range jobSettings {
		healthcheck := jobSetting.MustAsHealthcheck()
		healthcheckData := &queue.HealthcheckExecData{
			HealthcheckSetting: jobSetting,
			Healthcheck:        healthcheck,
			Project:            jobSetting.BelongToProject,
			App:                jobSetting.BelongToApp,
			Task: &entity.Task{
				ID:       gofn.Must(ulid.NewStringULID()),
				Scope:    jobSetting.Scope,
				ObjectID: jobSetting.ObjectID,
				TargetID: jobSetting.ID,
				Type:     base.TaskTypeHealthcheck,
				Status:   base.TaskStatusNotStarted,
				Config: entity.TaskConfig{
					MaxRetry:   healthcheck.MaxRetry,
					RetryDelay: healthcheck.RetryDelay,
					Timeout:    healthcheck.Timeout,
				},
				Version:   entity.CurrentTaskVersion,
				StartedAt: timeNow,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
			RefObjects:    baseData.RefObjects,
			NotifEventMap: baseData.NotifEventMap,
		}
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return q.doHealthcheckItem(ctx, healthcheck, healthcheckData, &savingTasks)
		})
	}

	// Execute all health check tasks concurrently
	_ = gofn.ExecTasksEx(ctx, 100, false, execFuncs...) //nolint:mnd

	// Save tasks in DB
	err = q.taskRepo.UpsertMulti(ctx, q.db, savingTasks,
		entity.TaskUpsertingConflictColsByUK, nil)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) doHealthcheckItem(
	ctx context.Context,
	healthcheck *entity.Healthcheck,
	healthcheckData *queue.HealthcheckExecData,
	savingTasks *[]*entity.Task,
) error {
	lockKey := fmt.Sprintf(taskHealthcheckLockKey, healthcheckData.HealthcheckSetting.ID)
	success, releaser, err := q.taskService.CreateRedisLock(ctx, lockKey, time.Minute)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !success {
		return nil
	}
	defer releaser()

	err = q.healthcheckExecutor(ctx, healthcheckData)
	if healthcheck.SaveResultTasks {
		*savingTasks = append(*savingTasks, healthcheckData.Task)
	}
	return apperrors.Wrap(err)
}

func (q *taskQueue) loadHealthcheckData(
	ctx context.Context,
	db database.IDB,
	taskData *queue.HealthcheckExecData,
) ([]*entity.Setting, error) {
	// Query items from cache first
	queryDB := false
	healthcheckSettings, err := q.healthcheckSettingsRepo.Get(ctx)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			queryDB = true
		} else {
			return nil, apperrors.Wrap(err)
		}
	}

	if queryDB {
		healthcheckSettings, err = q.loadHealthcheckDataFromDB(ctx, db)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	if healthcheckSettings == nil {
		return nil, nil
	}

	timeNowSecs := timeutil.NowUTC().Unix()
	validJobSettings := make([]*entity.Setting, 0, len(healthcheckSettings.Settings))
	for _, jobSetting := range healthcheckSettings.Settings {
		healthcheck := jobSetting.MustAsHealthcheck()
		interval := int64(healthcheck.Interval.ToDuration().Seconds())
		if timeNowSecs%interval < min(interval, 5) { //nolint:mnd
			validJobSettings = append(validJobSettings, jobSetting)
		}
	}
	if len(validJobSettings) == 0 {
		return nil, nil
	}

	// Load history notification events
	taskData.NotifEventMap, err = q.healthcheckNotifEventRepo.GetAll(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	taskData.RefObjects = healthcheckSettings.RefObjects

	return validJobSettings, nil
}

func (q *taskQueue) loadHealthcheckDataFromDB(
	ctx context.Context,
	db database.IDB,
) (*cacheentity.HealthcheckSettings, error) {
	dbSettings, _, err := q.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeHealthcheck),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectRelation("BelongToProject",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var validHealthchecks []*entity.Setting
	for _, healthcheck := range dbSettings {
		if healthcheck.BelongToApp != nil {
			healthcheck.BelongToProject = healthcheck.BelongToApp.Project
			healthcheck.BelongToApp.Project = nil // NOTE: we do this only because the app may be stored in redis
		}
		project := healthcheck.BelongToProject
		app := healthcheck.BelongToApp
		if app != nil && app.Status != base.AppStatusActive {
			continue
		}
		if project != nil && project.Status != base.ProjectStatusActive {
			continue
		}
		validHealthchecks = append(validHealthchecks, healthcheck)
	}

	// Load reference objects
	refObjects, err := q.settingService.LoadReferenceObjects(ctx, db, nil,
		true, false, validHealthchecks...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	healthcheckSettings := &cacheentity.HealthcheckSettings{
		Settings:   validHealthchecks,
		RefObjects: refObjects,
	}

	// Put data in cache
	err = q.healthcheckSettingsRepo.Set(ctx, healthcheckSettings, cacheHealthcheckSettingsExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return healthcheckSettings, nil
}
