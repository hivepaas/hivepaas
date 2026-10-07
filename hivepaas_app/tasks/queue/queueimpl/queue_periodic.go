package queueimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

const (
	taskPeriodicLockKey = "task:periodic:%v:lock"
	// periodicLockTTL is the least a periodic job's lock is held for; it is
	// held periodicLockMargin past a longer run's bound.
	periodicLockTTL          = time.Minute
	periodicLockMargin       = 10 * time.Second
	cachePeriodicSettingsExp = 5 * time.Minute
	defaultPeriodicBatchSize = 100
	// periodicRunSlots is how many runs go at once. A run takes a slot before
	// it starts, and a job due while they are all taken waits for one.
	periodicRunSlots = 100
)

func (q *taskQueue) RegisterPeriodicExecutor(execFunc queue.PeriodicExecFunc) {
	if !q.isWorkerMode() {
		return
	}
	q.periodicExecutor = execFunc
}

// doPeriodicJob is a tick: the jobs due are started, in the background. The
// tick does not wait for them - one that hangs, on a service that does not
// answer, would hold up every other job as long as it hangs.
func (q *taskQueue) doPeriodicJob(
	ctx context.Context,
) error {
	if q.periodicExecutor == nil {
		return hperrors.NewUnavailable("Task executor function for periodic jobs")
	}

	baseData := &queue.PeriodicExecData{}
	jobSettings, err := q.loadPeriodicJobData(ctx, q.db, baseData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	for _, jobSetting := range jobSettings {
		periodicData, err := periodicExecDataOf(jobSetting, baseData, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
		q.startPeriodicRun(ctx, periodicData)
	}
	return nil
}

// startPeriodicRun runs a due job in the background, once it has a slot.
func (q *taskQueue) startPeriodicRun(ctx context.Context, periodicData *queue.PeriodicExecData) {
	q.periodicRuns.Add(1)
	safego.GoWithLogger(q.logger, "taskQueue.periodicRun", func() {
		defer q.periodicRuns.Done()
		select {
		case q.periodicSlots <- struct{}{}:
		case <-ctx.Done():
			return
		}
		defer func() { <-q.periodicSlots }()
		q.notePeriodicResult(periodicData.PeriodicSetting.ID, q.doPeriodicTask(ctx, periodicData))
	})
}

// notePeriodicResult logs how a run went when that changed: once when a job
// starts failing, with the error, and once when it runs again - not a line a
// run, every 15 seconds, while a server it needs is down.
func (q *taskQueue) notePeriodicResult(jobID string, err error) {
	q.periodicErrMu.Lock()
	defer q.periodicErrMu.Unlock()
	if q.periodicErrs == nil {
		q.periodicErrs = map[string]string{}
	}
	last := q.periodicErrs[jobID]
	switch {
	case err == nil && last != "":
		delete(q.periodicErrs, jobID)
		q.logger.Infof("periodic job %s runs again", jobID)
	case err != nil && err.Error() != last:
		q.periodicErrs[jobID] = err.Error()
		q.logger.Errorf("periodic job %s: %v", jobID, err)
	}
}

// periodicExecDataOf is what a due job's run is given. Its RefObjects is a copy
// of the round's: the runs go concurrently, and what they load - a
// notification's settings, for one - goes into it. Two runs writing into the
// round's maps at once is a fatal error, not a panic: it ends the process.
func periodicExecDataOf(
	jobSetting *entity.Setting,
	round *queue.PeriodicExecData,
	timeNow time.Time,
) (*queue.PeriodicExecData, error) {
	periodicJob := jobSetting.MustAsPeriodicJob()
	scope, err := round.RefObjects.GetObjectScope(jobSetting.Scope, jobSetting.ObjectID, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refObjects := entity.NewRefObjects()
	refObjects.AddRefObjects(round.RefObjects)
	return &queue.PeriodicExecData{
		PeriodicSetting: jobSetting,
		Scope:           scope,
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
		RefObjects: refObjects,
	}, nil
}

// doPeriodicTask is one run of a job: under the job's lock, within its
// timeout or its type's ceiling, and saved when the executor says so.
func (q *taskQueue) doPeriodicTask(
	ctx context.Context,
	periodicData *queue.PeriodicExecData,
) error {
	// The lock outlives the run's bound, so that no other worker starts the
	// job while this run may still be going.
	timeout := resolveTaskTimeout(periodicData.Task)
	lockKey := fmt.Sprintf(taskPeriodicLockKey, periodicData.PeriodicSetting.ID)
	success, releaser, err := q.taskService.CreateRedisLock(ctx, lockKey,
		max(periodicLockTTL, timeout+periodicLockMargin))
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !success {
		return nil
	}
	defer releaser()

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = q.periodicExecutor(runCtx, periodicData)
	if periodicData.SaveTask {
		err = errors.Join(err, q.taskRepo.UpsertMulti(ctx, q.db, []*entity.Task{periodicData.Task},
			entity.TaskUpsertingConflictColsByUK, nil))
	}
	return hperrors.Wrap(err)
}

// periodicCache is the active periodic jobs and what they refer to, read from
// the database every cachePeriodicSettingsExp, and when a setting they depend
// on changes.
type periodicCache struct {
	settingsMap map[string]*entity.Setting
	refObjects  *entity.RefObjects
	lastLoaded  time.Time
}

// intervalOf is a job's interval, in seconds; 0 for a job not in the cache.
func (c *periodicCache) intervalOf(id string) int64 {
	if c == nil {
		return 0
	}
	setting := c.settingsMap[id]
	if setting == nil {
		return 0
	}
	return int64(setting.MustAsPeriodicJob().Interval.ToDuration().Seconds())
}

// loadPeriodicJobData is the jobs due now, their next run scheduled, with the
// cache's objects in taskData for their runs.
func (q *taskQueue) loadPeriodicJobData(
	ctx context.Context,
	db database.IDB,
	taskData *queue.PeriodicExecData,
) ([]*entity.Setting, error) {
	q.periodicCacheMu.RLock()
	cache := q.periodicCache
	q.periodicCacheMu.RUnlock()

	needReload := cache == nil || time.Since(cache.lastLoaded) > cachePeriodicSettingsExp
	// Every event waiting is one reload: what changed since the last is read
	// once, not once a tick for as many events as came in.
	for drained := false; !drained && q.periodicReloadChan != nil; {
		select {
		case _, ok := <-q.periodicReloadChan:
			if !ok {
				q.periodicReloadChan = nil
			}
			needReload = true
		default:
			drained = true
		}
	}

	if needReload {
		var err error
		cache, err = q.loadPeriodicJobDataFromDB(ctx, db, cache)
		if err != nil {
			return nil, hperrors.Wrap(err)
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
	dueJobs, err := q.periodicSettingsRepo.GetDueJobs(ctx, timeNowSecs, int64(batchSize))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(dueJobs) == 0 {
		return nil, nil
	}

	validJobSettings := make([]*entity.Setting, 0, len(dueJobs))
	nextRuns := make(map[string]int64, len(dueJobs))
	for _, due := range dueJobs {
		jobSetting, exists := cache.settingsMap[due.ID]
		if !exists {
			_ = q.periodicSettingsRepo.RemoveJob(ctx, due.ID)
			continue
		}
		periodic := jobSetting.MustAsPeriodicJob()
		interval := int64(periodic.Interval.ToDuration().Seconds())
		if interval <= 0 {
			_ = q.periodicSettingsRepo.RemoveJob(ctx, due.ID)
			continue
		}
		nextRuns[due.ID] = nextRunAfter(due.DueSecs, timeNowSecs, interval)
		validJobSettings = append(validJobSettings, jobSetting)
	}
	// Scheduled before they run, all in one call. Not scheduled, they are not
	// run either: they would be due again next tick, and run again.
	if err := q.periodicSettingsRepo.ScheduleJobs(ctx, nextRuns, false); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(validJobSettings) == 0 {
		return nil, nil
	}

	taskData.RefObjects = cache.refObjects

	return validJobSettings, nil
}

// nextRunAfter is when a job due at dueSecs runs next: an interval after it
// was due, not after it was seen - seen is up to a tick late, and a cadence
// counted from there drifts by that much a run. After a stall that is past
// already: then an interval from now, the runs missed not made up.
func nextRunAfter(dueSecs, nowSecs, intervalSecs int64) int64 {
	next := dueSecs + intervalSecs
	if next <= nowSecs {
		next = nowSecs + intervalSecs
	}
	return next
}

// loadPeriodicJobDataFromDB reads the active jobs and what they refer to, and
// puts into the schedule the jobs that are not in it.
//
// The schedule in redis outlives a reload. Every reload used to give every
// job a new slot, as if new - and with one reload every five minutes, a job
// on an interval longer than that was pushed on before it was ever due, and
// never ran. A job is given a slot when it is new to the schedule, and again
// when its interval changed since the previous cache.
func (q *taskQueue) loadPeriodicJobDataFromDB(
	ctx context.Context,
	db database.IDB,
	previous *periodicCache,
) (*periodicCache, error) {
	dbSettings, _, err := q.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypePeriodicJob),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refIDs := &entity.RefObjectIDs{}
	for _, setting := range dbSettings {
		refIDs.AddScopeObjectIDOfSettings(setting)
		rIDs, err := setting.GetRefObjectIDs()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		refIDs.AddRefIDs(rIDs)
	}

	// Load reference objects
	refObjects := entity.NewRefObjects()
	err = q.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, true, refIDs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	settingsMap := make(map[string]*entity.Setting, len(dbSettings))
	timeNowSecs := timeutil.NowUTC().Unix()
	kept := make(map[string]int64, len(dbSettings))
	changed := make(map[string]int64)
	for _, setting := range dbSettings {
		scope, _ := refObjects.GetObjectScope(setting.Scope, setting.ObjectID, true)
		if scope == nil {
			continue
		}
		settingsMap[setting.ID] = setting
		periodic := setting.MustAsPeriodicJob()
		interval := int64(periodic.Interval.ToDuration().Seconds())
		if interval <= 0 {
			continue
		}
		// A slot within the interval, the jobs spread over it.
		slot := timeNowSecs + int64(stringHash(setting.ID)%uint64(interval)) //nolint:gosec
		if before := previous.intervalOf(setting.ID); before != 0 && before != interval {
			changed[setting.ID] = slot
		} else {
			kept[setting.ID] = slot
		}
	}
	if err := q.periodicSettingsRepo.ScheduleJobs(ctx, kept, true); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := q.periodicSettingsRepo.ScheduleJobs(ctx, changed, false); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &periodicCache{
		settingsMap: settingsMap,
		refObjects:  refObjects,
		lastLoaded:  time.Now(),
	}, nil
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
