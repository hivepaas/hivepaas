package queueimpl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

const (
	taskPeriodicLockKey      = "task:periodic:%v:lock"
	cachePeriodicSettingsExp = 5 * time.Minute
	defaultPeriodicBatchSize = 100
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
			Scope:           getObjectScope(jobSetting, baseData.RefObjects),
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

type periodicCache struct {
	settingsMap map[string]*entity.Setting
	refObjects  *entity.RefObjects
	lastLoaded  time.Time
}

func (q *taskQueue) loadPeriodicJobData(
	ctx context.Context,
	db database.IDB,
	taskData *queue.PeriodicExecData,
) ([]*entity.Setting, error) {
	// Check if reload is needed
	needReload := false
	q.periodicCacheMu.RLock()
	cache := q.periodicCache
	q.periodicCacheMu.RUnlock()

	if cache == nil || time.Since(cache.lastLoaded) > cachePeriodicSettingsExp {
		needReload = true
	} else if q.periodicReloadChan != nil {
		select {
		case <-q.periodicReloadChan:
			needReload = true
		default:
		}
	}

	if needReload {
		var err error
		cache, err = q.loadPeriodicJobDataFromDB(ctx, db)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		q.periodicCacheMu.Lock()
		q.periodicCache = cache
		q.periodicCacheMu.Unlock()
	}

	if cache == nil || len(cache.settingsMap) == 0 {
		return nil, nil
	}

	timeNowSecs := timeutil.NowUTC().Unix()
	batchSize := q.periodicBatchSize
	if batchSize <= 0 {
		batchSize = defaultPeriodicBatchSize
	}
	dueJobIDs, err := q.periodicSettingsRepo.GetDueJobIDs(ctx, timeNowSecs, int64(batchSize))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(dueJobIDs) == 0 {
		return nil, nil
	}

	validJobSettings := make([]*entity.Setting, 0, len(dueJobIDs))
	for _, dueJobID := range dueJobIDs {
		jobSetting, exists := cache.settingsMap[dueJobID]
		if !exists {
			_ = q.periodicSettingsRepo.RemoveJob(ctx, dueJobID)
			continue
		}
		periodic := jobSetting.MustAsPeriodicJob()
		interval := int64(periodic.Interval.ToDuration().Seconds())
		if interval <= 0 {
			_ = q.periodicSettingsRepo.RemoveJob(ctx, dueJobID)
			continue
		}

		// Reschedule next run time in ZSET
		nextRun := timeNowSecs + interval
		_ = q.periodicSettingsRepo.ScheduleJob(ctx, dueJobID, nextRun)

		validJobSettings = append(validJobSettings, jobSetting)
	}
	if len(validJobSettings) == 0 {
		return nil, nil
	}

	taskData.RefObjects = cache.refObjects

	return validJobSettings, nil
}

func (q *taskQueue) loadPeriodicJobDataFromDB(
	ctx context.Context,
	db database.IDB,
) (*periodicCache, error) {
	dbSettings, _, err := q.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypePeriodicJob),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	refIDs := &entity.RefObjectIDs{}
	for _, setting := range dbSettings {
		switch setting.Scope {
		case base.ObjectScopeGlobal:
		case base.ObjectScopeApp:
			refIDs.RefAppIDs = append(refIDs.RefAppIDs, setting.ObjectID)
		case base.ObjectScopeProject:
			refIDs.RefProjectIDs = append(refIDs.RefProjectIDs, setting.ObjectID)
		case base.ObjectScopeProjectEnv:
			refIDs.RefProjectEnvIDs = append(refIDs.RefProjectEnvIDs, setting.ObjectID)
		case base.ObjectScopeUser:
			refIDs.RefUserIDs = append(refIDs.RefUserIDs, setting.ObjectID)
		}
		rIDs, err := setting.GetRefObjectIDs()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		refIDs.AddRefIDs(rIDs)
	}

	// Load reference objects
	refObjects := entity.NewRefObjects()
	err = q.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, true, refIDs)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settingsMap := make(map[string]*entity.Setting, len(dbSettings))
	timeNowSecs := timeutil.NowUTC().Unix()
	for _, setting := range dbSettings {
		scope := getObjectScope(setting, refObjects)
		if scope != nil {
			settingsMap[setting.ID] = setting
			periodic := setting.MustAsPeriodicJob()
			interval := int64(periodic.Interval.ToDuration().Seconds())
			if interval <= 0 {
				continue
			}
			slot := int64(stringHash(setting.ID) % uint64(interval)) //nolint:gosec
			initialRun := timeNowSecs + slot
			_ = q.periodicSettingsRepo.ScheduleJob(ctx, setting.ID, initialRun)
		}
	}

	return &periodicCache{
		settingsMap: settingsMap,
		refObjects:  refObjects,
		lastLoaded:  time.Now(),
	}, nil
}

func getObjectScope(setting *entity.Setting, refObjects *entity.RefObjects) *entity.ObjectScope {
	switch setting.Scope {
	case base.ObjectScopeApp:
		app := refObjects.RefApps[setting.ObjectID]
		if (app == nil || app.Status != base.AppStatusActive) ||
			(app.ProjectEnv == nil || app.ProjectEnv.Status != base.ProjectStatusActive) ||
			(app.Project == nil || app.Project.Status != base.ProjectStatusActive) {
			return nil
		}
		return app.GetObjectScope()
	case base.ObjectScopeProject:
		project := refObjects.RefProjects[setting.ObjectID]
		if project == nil || project.Status != base.ProjectStatusActive {
			return nil
		}
		return project.GetObjectScope()
	case base.ObjectScopeProjectEnv:
		projectEnv := refObjects.RefProjectEnvs[setting.ObjectID]
		if (projectEnv == nil || projectEnv.Status != base.ProjectStatusActive) ||
			(projectEnv.Project == nil || projectEnv.Project.Status != base.ProjectStatusActive) {
			return nil
		}
		return projectEnv.GetObjectScope()
	case base.ObjectScopeUser:
		user := refObjects.RefUsers[setting.ObjectID]
		if user == nil || user.Status != base.UserStatusActive || user.IsAccessExpired() {
			return nil
		}
		return user.GetObjectScope()
	case base.ObjectScopeGlobal:
		return entity.NewObjectScopeGlobal()
	}
	return nil
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
