package queue

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	taskHealthcheckLockKey      = "task:healthcheck:lock"
	taskHealthcheckLockMaxRetry = 3
	cacheHealthcheckSettingsExp = 5 * time.Minute
)

type HealthcheckExecData struct {
	HealthcheckSetting *entity.Setting
	Healthcheck        *entity.Healthcheck
	Task               *entity.Task
	Project            *entity.Project
	App                *entity.App

	// RefObjects can be used as a cache to store objects
	RefObjects    *entity.RefObjects
	NotifEventMap map[string]*cacheentity.HealthcheckNotifEvent
}

func (t *HealthcheckExecData) AddRefObjects(refObjects *entity.RefObjects) {
	if t.RefObjects == nil {
		t.RefObjects = refObjects
	} else {
		t.RefObjects.AddRefObjects(refObjects)
	}
}

type HealthcheckExecFunc func(context.Context, *HealthcheckExecData) error

func (q *taskQueue) RegisterHealthcheckExecutor(execFunc HealthcheckExecFunc) {
	if !q.isWorkerMode() {
		return
	}
	q.healthcheckExecutor = execFunc
}

func (q *taskQueue) doHealthcheck(
	ctx context.Context,
) error {
	// Make sure only one worker processes this task at a time
	success, _, err := q.healthcheckTaskLock(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !success { // another worker is doing this task
		return nil
	}

	executorFunc := q.healthcheckExecutor
	if executorFunc == nil {
		return apperrors.NewUnavailable("Task executor function for healthcheck")
	}

	baseData := &HealthcheckExecData{}
	jobSettings, err := q.loadHealthcheckData(ctx, q.db, baseData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	savingTasks := make([]*entity.Task, 0, len(jobSettings))
	execFuncs := make([]func(ctx context.Context) error, 0, len(jobSettings))

	for _, jobSetting := range jobSettings {
		healthcheck := jobSetting.MustAsHealthcheck()
		healthcheckData := &HealthcheckExecData{
			HealthcheckSetting: jobSetting,
			Healthcheck:        healthcheck,
			Project:            jobSetting.BelongToProject,
			App:                jobSetting.BelongToApp,
			Task: &entity.Task{
				ID:       gofn.Must(ulid.NewStringULID()),
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
		if healthcheck.SaveResultTasks {
			savingTasks = append(savingTasks, healthcheckData.Task)
		}
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return executorFunc(ctx, healthcheckData) //nolint:wrapcheck
		})
	}

	// Execute all health check tasks concurrently
	_ = gofn.ExecTasksEx(ctx, 20, false, execFuncs...) //nolint:mnd

	// Save tasks in DB
	err = q.taskRepo.UpsertMulti(ctx, q.db, savingTasks,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) healthcheckTaskLock(ctx context.Context) (bool, func(), error) {
	interval := config.Current.Tasks.Healthcheck.BaseInterval
	retries := 0
	wait := time.Duration(0)
	for {
		success, releaser, err := q.taskService.CreateRedisLock(ctx, taskHealthcheckLockKey, interval-time.Second)
		if err != nil {
			if retries >= taskHealthcheckLockMaxRetry {
				return false, nil, apperrors.Wrap(err)
			}
			retries++
			wait += time.Second
			time.Sleep(wait)
			continue
		}
		return success, releaser, nil
	}
}

func (q *taskQueue) loadHealthcheckData(
	ctx context.Context,
	db database.IDB,
	taskData *HealthcheckExecData,
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
		if timeNowSecs%int64(healthcheck.Interval.ToDuration().Seconds()) > 5 { //nolint:mnd
			continue
		}
		validJobSettings = append(validJobSettings, jobSetting)
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
