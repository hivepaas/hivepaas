package gocronqueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const testTaskType base.TaskType = "task:test"

// serverRunning is a server with gocron started, whose test tasks say they ran.
func serverRunning(t *testing.T) (*Server, <-chan string) {
	t.Helper()
	sched, err := gocron.NewScheduler(gocron.WithLimitConcurrentJobs(2, gocron.LimitModeWait))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sched.Shutdown() })
	ran := make(chan string, 10)
	server := &Server{
		config: &Config{Logger: &lines{}, TaskMap: map[base.TaskType]TaskExecFunc{
			testTaskType: func(taskID, _ string) time.Time {
				ran <- taskID
				return time.Time{}
			},
		}},
		scheduler:         sched,
		jobMap:            make(map[string]*jobData),
		startedAt:         time.Now(),
		pauseLimit:        300 * time.Millisecond,
		pauseRecheckEvery: 20 * time.Millisecond,
	}
	sched.Start()
	return server, ran
}

func dueTask(id string) *entity.Task {
	return &entity.Task{ID: id, Type: testTaskType, RunAt: time.Now()}
}

func expectRun(t *testing.T, ran <-chan string, id string, within time.Duration) {
	t.Helper()
	select {
	case got := <-ran:
		assert.Equal(t, id, got)
	case <-time.After(within):
		t.Fatalf("%s did not run within %v", id, within)
	}
}

func expectNoRun(t *testing.T, ran <-chan string, during time.Duration) {
	t.Helper()
	select {
	case got := <-ran:
		t.Fatalf("%s ran while the scheduler was paused", got)
	case <-time.After(during):
	}
}

// What the pauses below hold: a task given to a running scheduler runs at once.
func TestATaskGivenToARunningSchedulerRuns(t *testing.T) {
	server, ran := serverRunning(t)

	assert.NoError(t, server.ScheduleTask(context.Background(), dueTask("t1")))

	expectRun(t, ran, "t1", time.Second)
}

// A stop is a pause for a restart: a task due meanwhile waits. One that reaches a
// process the restart does not replace - a worker when only the app is changed,
// the app when an update failed before it was scaled down - held its scheduler
// until something restarted it; the pause ends on its own now, and the tasks it
// held run.
func TestAPauseHoldsTasksUntilItEndsOnItsOwn(t *testing.T) {
	server, ran := serverRunning(t)

	assert.NoError(t, server.StopScheduler())
	assert.NoError(t, server.ScheduleTask(context.Background(), dueTask("t1")))

	expectNoRun(t, ran, 150*time.Millisecond)
	expectRun(t, ran, "t1", 2*time.Second)
	logs := server.config.Logger.(*lines)
	assert.True(t, logs.has("resumed on its own"), logs.logs)
}

// A start ends the pause before its limit, and the tasks it held run.
func TestAStartRunsTheTasksAPauseHeld(t *testing.T) {
	server, ran := serverRunning(t)
	server.pauseLimit = time.Hour

	assert.NoError(t, server.StopScheduler())
	assert.NoError(t, server.ScheduleTask(context.Background(), dueTask("t1")))
	expectNoRun(t, ran, 100*time.Millisecond)
	assert.NoError(t, server.StartScheduler())

	expectRun(t, ran, "t1", time.Second)
}

// slowScheduler hands back the first job it makes only once that job has had
// time to run: a job due at once can run before NewJob returns.
type slowScheduler struct {
	gocron.Scheduler
	made atomic.Int32
}

func (s *slowScheduler) NewJob(def gocron.JobDefinition, task gocron.Task, opts ...gocron.JobOption) (gocron.Job, error) {
	job, err := s.Scheduler.NewJob(def, task, opts...)
	if s.made.Add(1) == 1 {
		time.Sleep(100 * time.Millisecond)
	}
	return job, err
}

// A task scheduled while paused can be held - by its own job, rescheduling it -
// before the scheduling of it is done. That job was then put back as the task's,
// in place of the one holding it, and the task was lost: its job had run, and a
// scan finding the task again took it for scheduled.
func TestATaskHeldBeforeItsSchedulingEndsIsNotLost(t *testing.T) {
	server, ran := serverRunning(t)
	server.scheduler = &slowScheduler{Scheduler: server.scheduler}
	server.pauseLimit = time.Hour
	task := dueTask("t1")

	assert.NoError(t, server.StopScheduler())
	assert.NoError(t, server.ScheduleTask(context.Background(), task))
	assert.NoError(t, server.StartScheduler())
	assert.NoError(t, server.ScheduleTask(context.Background(), task)) // a scan finding it again

	expectRun(t, ran, "t1", time.Second)
}
