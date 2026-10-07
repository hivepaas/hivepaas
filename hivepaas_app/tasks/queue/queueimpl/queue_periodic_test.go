package queueimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// lockService grants every lock, and keeps how long it was asked for.
type lockService struct {
	taskservice.Service
	exp time.Duration
}

func (f *lockService) CreateRedisLock(_ context.Context, _ string, exp time.Duration) (bool, func(), error) {
	f.exp = exp
	return true, func() {}, nil
}

// periodicJobs is the periodic jobs the database has, and how often they
// were read.
type periodicJobs struct {
	repository.SettingRepo
	settings []*entity.Setting
	reads    int
}

func (r *periodicJobs) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	r.reads++
	return r.settings, nil, nil
}

// refLoader gives the jobs what they refer to: one app, active in an active
// project and environment.
type refLoader struct{ settingservice.Service }

func (refLoader) LoadRefObjectsByIDsSkipMissing(_ context.Context, _ database.IDB, refObjects **entity.RefObjects,
	_ *entity.ObjectScope, _ bool, _ *entity.RefObjectIDs) error {
	(*refObjects).RefApps["app"] = &entity.App{ID: "app", Status: base.AppStatusActive,
		ProjectID: "p", ProjectEnvID: "p:dev"}
	(*refObjects).RefProjects["p"] = &entity.Project{ID: "p", Status: base.ProjectStatusActive}
	(*refObjects).RefProjectEnvs["p:dev"] = &entity.ProjectEnv{ID: "p:dev", Status: base.ProjectStatusActive}
	return nil
}

// schedule is the schedule in redis: when each job runs next, and the jobs
// due now.
type schedule struct {
	cacherepository.PeriodicSettingsRepo
	next map[string]int64
	due  []cacherepository.DueJob
}

func (s *schedule) GetDueJobs(context.Context, int64, int64) ([]cacherepository.DueJob, error) {
	return s.due, nil
}

func (s *schedule) ScheduleJobs(_ context.Context, nextRuns map[string]int64, keepExisting bool) error {
	for id, at := range nextRuns {
		if _, there := s.next[id]; there && keepExisting {
			continue
		}
		s.next[id] = at
	}
	return nil
}

func (s *schedule) RemoveJob(_ context.Context, id string) error {
	delete(s.next, id)
	return nil
}

func healthcheckJob(t *testing.T, id string, interval time.Duration) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: base.SettingTypePeriodicJob, Scope: base.ObjectScopeApp,
		ObjectID: "app", Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(&entity.PeriodicJob{Interval: timeutil.Duration(interval)}))
	return setting
}

// periodicQueue is a worker's queue over the jobs and the schedule given,
// with exec as every job's run.
func periodicQueue(jobs *periodicJobs, sched *schedule, exec queue.PeriodicExecFunc) (*taskQueue, *lockService) {
	locks := &lockService{}
	return &taskQueue{
		logger: logging.GlobalLogger(), settingRepo: jobs, settingService: refLoader{},
		periodicSettingsRepo: sched, taskService: locks, periodicExecutor: exec,
		periodicSlots: make(chan struct{}, periodicRunSlots),
	}, locks
}

func periodicRun(t *testing.T, timeout time.Duration, exec queue.PeriodicExecFunc) *lockService {
	t.Helper()
	q, locks := periodicQueue(nil, nil, exec)
	data := &queue.PeriodicExecData{
		PeriodicSetting: &entity.Setting{ID: "job"},
		Task: &entity.Task{Type: base.TaskTypePeriodicExec,
			Config: entity.TaskConfig{Timeout: timeutil.Duration(timeout)}},
	}
	_ = q.doPeriodicTask(context.Background(), data)
	return locks
}

// A periodic run is bounded by its job's timeout, so that one that hangs does
// not go on for as long as it hangs; its lock outlives the bound.
func TestPeriodicRunIsBoundedByItsTimeout(t *testing.T) {
	var deadline time.Time
	var hasDeadline bool
	locks := periodicRun(t, 30*time.Second, func(ctx context.Context, _ *queue.PeriodicExecData) error {
		deadline, hasDeadline = ctx.Deadline()
		return nil
	})
	if assert.True(t, hasDeadline) {
		assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, 5*time.Second)
	}
	assert.Equal(t, time.Minute, locks.exp, "a minute at least")

	locks = periodicRun(t, 5*time.Minute, func(context.Context, *queue.PeriodicExecData) error { return nil })
	assert.Equal(t, 5*time.Minute+periodicLockMargin, locks.exp, "past a longer bound")

	ended := make(chan struct{})
	go func() {
		periodicRun(t, 50*time.Millisecond, func(ctx context.Context, _ *queue.PeriodicExecData) error {
			<-ctx.Done() // a run that hangs until it is stopped
			return ctx.Err()
		})
		close(ended)
	}()
	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("a hanging run was not stopped at its timeout")
	}
}

// The runs of a round go concurrently, and what they load - a notification's
// settings among them - goes into their RefObjects: each has its own, filled
// from the round's, so that two runs never write into one map. Writing into
// one map from two goroutines is a fatal error: it took the backend down.
func TestEachPeriodicRunHasRefObjectsOfItsOwn(t *testing.T) {
	round := &queue.PeriodicExecData{RefObjects: entity.NewRefObjects()}
	round.RefObjects.RefApps["app"] = &entity.App{ID: "app", ProjectEnv: &entity.ProjectEnv{},
		Project: &entity.Project{}}
	round.RefObjects.RefSettings["shared"] = &entity.Setting{ID: "shared"}

	one, err := periodicExecDataOf(healthcheckJob(t, "one", time.Minute), round, time.Now())
	assert.NoError(t, err)
	two, err := periodicExecDataOf(healthcheckJob(t, "two", time.Minute), round, time.Now())
	assert.NoError(t, err)

	one.RefObjects.RefSettings["loaded"] = &entity.Setting{ID: "loaded"}
	assert.NotContains(t, two.RefObjects.RefSettings, "loaded")
	assert.NotContains(t, round.RefObjects.RefSettings, "loaded", "the round's, kept for the next")
	assert.Contains(t, two.RefObjects.RefSettings, "shared", "what the round loaded, each has")
	assert.Equal(t, "app", one.Scope.App.ID)
}

// The schedule in redis outlives a reload of the cache. Every reload used to
// give every job a new slot, as if new: with one reload every five minutes, a
// job on an interval longer than that was pushed on before it was ever due,
// and never ran.
func TestAReloadKeepsTheScheduleOfTheJobsThatDidNotChange(t *testing.T) {
	ctx := context.Background()
	jobs := &periodicJobs{settings: []*entity.Setting{healthcheckJob(t, "hourly", time.Hour)}}
	sched := &schedule{next: map[string]int64{}}
	q, _ := periodicQueue(jobs, sched, nil)

	first, err := q.loadPeriodicJobDataFromDB(ctx, nil, nil)
	assert.NoError(t, err)
	at := sched.next["hourly"]
	assert.NotZero(t, at, "a job new to the schedule is given a slot")

	// Reloaded after the job ran and was given its next run: it keeps it.
	ran := at + 3600
	sched.next["hourly"] = ran
	second, err := q.loadPeriodicJobDataFromDB(ctx, nil, first)
	assert.NoError(t, err)
	assert.Equal(t, ran, sched.next["hourly"])

	// Lost from redis - flushed, restarted - it is put back.
	delete(sched.next, "hourly")
	_, err = q.loadPeriodicJobDataFromDB(ctx, nil, second)
	assert.NoError(t, err)
	assert.NotZero(t, sched.next["hourly"])

	// Its interval shortened: it is given a slot on the new one, not left
	// waiting for the old.
	jobs.settings = []*entity.Setting{healthcheckJob(t, "hourly", 10*time.Second)}
	_, err = q.loadPeriodicJobDataFromDB(ctx, nil, second)
	assert.NoError(t, err)
	assert.LessOrEqual(t, sched.next["hourly"], time.Now().Unix()+10)
}

// The events waiting when a tick comes are one reload: what changed since the
// last is read once, not once a tick for as many events as came in.
func TestTheReloadEventsWaitingAreOneReload(t *testing.T) {
	jobs := &periodicJobs{settings: []*entity.Setting{healthcheckJob(t, "job", time.Minute)}}
	sched := &schedule{next: map[string]int64{}}
	q, _ := periodicQueue(jobs, sched, nil)
	events := make(chan *entity.SystemEvent, 8)
	q.periodicReloadChan = events
	q.periodicCache = &periodicCache{settingsMap: map[string]*entity.Setting{},
		refObjects: entity.NewRefObjects(), lastLoaded: time.Now()}
	for range 3 {
		events <- &entity.SystemEvent{}
	}

	for range 3 {
		_, err := q.loadPeriodicJobData(context.Background(), nil, &queue.PeriodicExecData{})
		assert.NoError(t, err)
	}

	assert.Equal(t, 1, jobs.reads)
}

// A job runs an interval after it was due, not after it was seen - seen is up
// to a tick late, and a cadence counted from there drifts by that much a run.
// After a stall that is past already: then an interval from now, the runs
// missed not made up.
func TestTheNextRunKeepsTheCadence(t *testing.T) {
	assert.Equal(t, int64(1010), nextRunAfter(1000, 1001, 10))
	assert.Equal(t, int64(1070), nextRunAfter(1000, 1060, 10))
}

// A tick does not wait for the runs it starts: one that hangs, on a service
// that does not answer, would hold up every other job as long as it hangs.
func TestATickDoesNotWaitForItsRuns(t *testing.T) {
	due := time.Now().Unix()
	jobs := &periodicJobs{settings: []*entity.Setting{healthcheckJob(t, "slow", time.Minute)}}
	sched := &schedule{next: map[string]int64{}, due: []cacherepository.DueJob{{ID: "slow", DueSecs: due}}}
	started := make(chan struct{})
	release := make(chan struct{})
	q, _ := periodicQueue(jobs, sched, func(context.Context, *queue.PeriodicExecData) error {
		close(started)
		<-release
		return nil
	})

	ticked := make(chan error, 1)
	go func() { ticked <- q.doPeriodicJob(context.Background()) }()
	select {
	case err := <-ticked:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the tick waited for the run")
	}
	<-started
	close(release)
	q.periodicRuns.Wait()

	assert.Equal(t, due+60, sched.next["slow"], "its next run, an interval after it was due")
}
