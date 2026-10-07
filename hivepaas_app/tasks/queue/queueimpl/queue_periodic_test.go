package queueimpl

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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

func periodicRun(t *testing.T, timeout time.Duration, exec queue.PeriodicExecFunc) *lockService {
	t.Helper()
	locks := &lockService{}
	q := &taskQueue{taskService: locks, periodicExecutor: exec}
	data := &queue.PeriodicExecData{
		PeriodicSetting: &entity.Setting{ID: "job"},
		Task: &entity.Task{Type: base.TaskTypePeriodicExec,
			Config: entity.TaskConfig{Timeout: timeutil.Duration(timeout)}},
	}
	var mu sync.Mutex
	var saving []*entity.Task
	_ = q.doPeriodicTask(context.Background(), data, &saving, &mu)
	return locks
}

// A periodic run is bounded by its job's timeout, so that one that hangs does
// not hold up the round - every other periodic job - for as long as it hangs;
// its lock outlives the bound.
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
	job := func(id string) *entity.Setting {
		setting := &entity.Setting{ID: id, Type: base.SettingTypePeriodicJob, Scope: base.ObjectScopeApp, ObjectID: "app"}
		assert.NoError(t, setting.SetData(&entity.PeriodicJob{}))
		return setting
	}

	one, err := periodicExecDataOf(job("one"), round, time.Now())
	assert.NoError(t, err)
	two, err := periodicExecDataOf(job("two"), round, time.Now())
	assert.NoError(t, err)

	one.RefObjects.RefSettings["loaded"] = &entity.Setting{ID: "loaded"}
	assert.NotContains(t, two.RefObjects.RefSettings, "loaded")
	assert.NotContains(t, round.RefObjects.RefSettings, "loaded", "the round's, kept for the next")
	assert.Contains(t, two.RefObjects.RefSettings, "shared", "what the round loaded, each has")
	assert.Equal(t, "app", one.Scope.App.ID)
}
