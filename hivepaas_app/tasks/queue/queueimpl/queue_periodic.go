package queueimpl

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	taskPeriodicLockKey      = "task:periodic:%v:lock"
	cachePeriodicSettingsExp = 5 * time.Minute
)

func (q *taskQueue) RegisterPeriodicExecutor(execFunc queue.PeriodicExecFunc) {
	if !q.isWorkerMode() {
		return
	}
	q.periodicExecutor = execFunc
}

func (q *taskQueue) doPeriodicJob(
	ctx context.Context,
) error {
	if q.periodicExecutor == nil {
		return apperrors.NewUnavailable("Task executor function for periodic jobs")
	}

	baseData := &queue.PeriodicExecData{}
	jobSettings, err := q.loadPeriodicJobData(ctx, q.db, baseData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if len(jobSettings) == 0 {
		return nil
	}

	var mu sync.Mutex
	timeNow := timeutil.NowUTC()
	savingTasks := make([]*entity.Task, 0, len(jobSettings))
	execFuncs := make([]func(ctx context.Context) error, 0, len(jobSettings))

	for _, jobSetting := range jobSettings {
		periodicJob := jobSetting.MustAsPeriodicJob()
		periodicData := &queue.PeriodicExecData{
			PeriodicSetting: jobSetting,
			Project:         jobSetting.BelongToProject,
			App:             jobSetting.BelongToApp,
			Task: &entity.Task{
				ID:       gofn.Must(ulid.NewStringULID()),
				Scope:    jobSetting.Scope,
				ObjectID: jobSetting.ObjectID,
				TargetID: jobSetting.ID,
				Type:     base.TaskTypePeriodicExec,
				Status:   base.TaskStatusNotStarted,
				Config: entity.TaskConfig{
					MaxRetry:   periodicJob.MaxRetry,
					RetryDelay: periodicJob.RetryDelay,
					Timeout:    periodicJob.Timeout,
				},
				Version:   entity.CurrentTaskVersion,
				RunAt:     timeNow,
				StartedAt: timeNow,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			},
			RefObjects: baseData.RefObjects,
		}
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return q.doPeriodicTask(ctx, periodicData, &savingTasks, &mu)
		})
	}

	// Execute all periodic tasks concurrently
	_ = gofn.ExecTasksEx(ctx, 100, false, execFuncs...) //nolint:mnd

	// Save tasks in DB
	err = q.taskRepo.UpsertMulti(ctx, q.db, savingTasks,
		entity.TaskUpsertingConflictColsByUK, nil)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) doPeriodicTask(
	ctx context.Context,
	periodicData *queue.PeriodicExecData,
	savingTasks *[]*entity.Task,
	mu *sync.Mutex,
) error {
	lockKey := fmt.Sprintf(taskPeriodicLockKey, periodicData.PeriodicSetting.ID)
	success, releaser, err := q.taskService.CreateRedisLock(ctx, lockKey, time.Minute)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !success {
		return nil
	}
	defer releaser()

	err = q.periodicExecutor(ctx, periodicData)
	if periodicData.SaveTask {
		mu.Lock()
		*savingTasks = append(*savingTasks, periodicData.Task)
		mu.Unlock()
	}

	return apperrors.Wrap(err)
}

func (q *taskQueue) loadPeriodicJobData(
	ctx context.Context,
	db database.IDB,
	taskData *queue.PeriodicExecData,
) ([]*entity.Setting, error) {
	// Query items from cache first
	queryDB := false
	periodicSettings, err := q.periodicSettingsRepo.Get(ctx)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			queryDB = true
		} else {
			return nil, apperrors.Wrap(err)
		}
	}

	if queryDB {
		periodicSettings, err = q.loadPeriodicJobDataFromDB(ctx, db)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	if periodicSettings == nil {
		return nil, nil
	}

	timeNowSecs := uint64(timeutil.NowUTC().Unix()) //nolint:gosec
	validJobSettings := make([]*entity.Setting, 0, len(periodicSettings.Settings))
	for _, jobSetting := range periodicSettings.Settings {
		periodic := jobSetting.MustAsPeriodicJob()
		interval := uint64(periodic.Interval.ToDuration().Seconds())
		if interval == 0 {
			continue
		}
		// Time-Bucket Jitter: Distribute execution slots evenly across the interval window
		// based on the deterministic hash of the setting ID to prevent thundering herd spikes.
		slot := stringHash(jobSetting.ID) % interval
		if timeNowSecs%interval == slot {
			validJobSettings = append(validJobSettings, jobSetting)
		}
	}
	if len(validJobSettings) == 0 {
		return nil, nil
	}

	taskData.RefObjects = periodicSettings.RefObjects

	return validJobSettings, nil
}

func (q *taskQueue) loadPeriodicJobDataFromDB(
	ctx context.Context,
	db database.IDB,
) (*cacheentity.PeriodicSettings, error) {
	dbSettings, _, err := q.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypePeriodicJob),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectRelation("BelongToProject",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToProjectEnv"),
		bunex.SelectRelation("BelongToProjectEnv.Project"),
		bunex.SelectRelation("BelongToApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp.ProjectEnv"),
		bunex.SelectRelation("BelongToApp.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var validSettings []*entity.Setting
	for _, setting := range dbSettings {
		if setting.BelongToApp != nil {
			setting.BelongToProject = setting.BelongToApp.Project
			setting.BelongToApp.Project = nil // NOTE: we do this only because the app may be stored in redis
			setting.BelongToProjectEnv = setting.BelongToApp.ProjectEnv
			setting.BelongToApp.ProjectEnv = nil
		} else if setting.BelongToProjectEnv != nil {
			setting.BelongToProject = setting.BelongToProjectEnv.Project
			setting.BelongToProjectEnv.Project = nil
		}
		project := setting.BelongToProject
		projectEnv := setting.BelongToProjectEnv
		app := setting.BelongToApp
		if app != nil && app.Status != base.AppStatusActive {
			continue
		}
		if projectEnv != nil && projectEnv.Status != base.ProjectStatusActive {
			continue
		}
		if project != nil && project.Status != base.ProjectStatusActive {
			continue
		}
		validSettings = append(validSettings, setting)
	}

	// Load reference objects
	refObjects := entity.NewRefObjects()
	err = q.settingService.LoadRefObjects(ctx, db, &refObjects, nil,
		true, false, validSettings...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	periodicSettings := &cacheentity.PeriodicSettings{
		Settings:   validSettings,
		RefObjects: refObjects,
	}

	// Put data in cache
	err = q.periodicSettingsRepo.Set(ctx, periodicSettings, cachePeriodicSettingsExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return periodicSettings, nil
}

// stringHash computes a deterministic 64-bit FNV-1a hash of a string for even time slot distribution.
func stringHash(s string) uint64 {
	var h uint64 = 14695981039346656037 // FNV offset basis
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211 // FNV prime
	}
	return h
}
