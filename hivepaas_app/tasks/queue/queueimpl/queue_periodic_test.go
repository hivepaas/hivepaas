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
