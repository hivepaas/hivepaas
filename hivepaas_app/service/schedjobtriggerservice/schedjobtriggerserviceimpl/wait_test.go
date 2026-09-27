package schedjobtriggerserviceimpl

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

// fakeRunRepo answers a list of tasks with their status as the test sets it.
type fakeRunRepo struct {
	repository.TaskRepo
	mu       sync.Mutex
	statuses map[string]base.TaskStatus
}

func (f *fakeRunRepo) set(id string, status base.TaskStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statuses[id] = status
}

func (f *fakeRunRepo) List(
	context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging, ...bunex.SelectQueryOption,
) ([]*entity.Task, *basedto.PagingMeta, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tasks := make([]*entity.Task, 0, len(f.statuses))
	for id, status := range f.statuses {
		tasks = append(tasks, &entity.Task{ID: id, Status: status})
	}
	return tasks, nil, nil
}

func waitService(t *testing.T, repo *fakeRunRepo) (*service, *[]string) {
	t.Helper()
	previous := config.Current()
	cfg := &config.Config{}
	cfg.Tasks.Triggers.WaitPollInterval = 10 * time.Millisecond
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })

	var canceled []string
	svc := &service{taskRepo: repo}
	svc.cancelRun = func(_ context.Context, taskID string) error {
		canceled = append(canceled, taskID)
		return nil
	}
	return svc, &canceled
}

func waitedRun(id, name string, timeout time.Duration) *schedjobtriggerservice.Run {
	return &schedjobtriggerservice.Run{Task: &entity.Task{ID: id}, JobName: name, Wait: true, Timeout: timeout}
}

func TestWaitForRunsGoesOnWhenEveryRunIsDone(t *testing.T) {
	repo := &fakeRunRepo{statuses: map[string]base.TaskStatus{
		"t1": base.TaskStatusNotStarted, "t2": base.TaskStatusNotStarted,
	}}
	svc, _ := waitService(t, repo)
	go func() {
		time.Sleep(30 * time.Millisecond)
		repo.set("t1", base.TaskStatusDone)
		time.Sleep(30 * time.Millisecond)
		repo.set("t2", base.TaskStatusDone)
	}()

	err := svc.WaitForRuns(context.Background(), []*schedjobtriggerservice.Run{
		waitedRun("t1", "migrate", time.Minute), waitedRun("t2", "seed", time.Minute),
	}, tasklog.NewNullStore())

	assert.NoError(t, err)
}

func TestWaitForRunsFailsWithARunThatFailed(t *testing.T) {
	repo := &fakeRunRepo{statuses: map[string]base.TaskStatus{"t1": base.TaskStatusFailed}}
	svc, _ := waitService(t, repo)

	err := svc.WaitForRuns(context.Background(), []*schedjobtriggerservice.Run{
		waitedRun("t1", "migrate", time.Minute),
	}, tasklog.NewNullStore())

	if assert.Error(t, err) {
		assert.Contains(t, hperrors.GetErrorDetail(err, ""), "migrate")
	}
}

func TestWaitForRunsFailsWithARunCanceled(t *testing.T) {
	repo := &fakeRunRepo{statuses: map[string]base.TaskStatus{"t1": base.TaskStatusCanceled}}
	svc, _ := waitService(t, repo)

	err := svc.WaitForRuns(context.Background(), []*schedjobtriggerservice.Run{
		waitedRun("t1", "migrate", time.Minute),
	}, tasklog.NewNullStore())

	assert.Error(t, err)
}

func TestWaitForRunsFailsPastARunsTimeout(t *testing.T) {
	repo := &fakeRunRepo{statuses: map[string]base.TaskStatus{"t1": base.TaskStatusNotStarted}}
	svc, _ := waitService(t, repo)

	err := svc.WaitForRuns(context.Background(), []*schedjobtriggerservice.Run{
		waitedRun("t1", "migrate", 50*time.Millisecond),
	}, tasklog.NewNullStore())

	if assert.Error(t, err) {
		assert.Contains(t, hperrors.GetErrorDetail(err, ""), "migrate")
	}
}

// A deploy canceled while it waits cancels the runs it waits for.
func TestWaitForRunsCancelsTheRunsWhenTheDeployIsCanceled(t *testing.T) {
	repo := &fakeRunRepo{statuses: map[string]base.TaskStatus{
		"t1": base.TaskStatusNotStarted, "t2": base.TaskStatusDone,
	}}
	svc, canceled := waitService(t, repo)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()

	err := svc.WaitForRuns(ctx, []*schedjobtriggerservice.Run{
		waitedRun("t1", "migrate", time.Minute), waitedRun("t2", "seed", time.Minute),
	}, tasklog.NewNullStore())

	assert.Error(t, err)
	assert.Equal(t, []string{"t1"}, *canceled, "only the run still going")
}

// Runs no trigger waits for are not waited for.
func TestWaitedRunsLeavesOutRunsNotWaitedFor(t *testing.T) {
	result := &schedjobtriggerservice.FireResult{Runs: []*schedjobtriggerservice.Run{
		{Task: &entity.Task{ID: "t1"}, Wait: true}, {Task: &entity.Task{ID: "t2"}},
	}}

	waited := result.WaitedRuns()

	if assert.Len(t, waited, 1) {
		assert.Equal(t, "t1", waited[0].Task.ID)
	}
}
