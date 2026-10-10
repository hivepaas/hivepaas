package docker

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

// An app never deployed has no service. Asked about it, Docker would read the
// empty filter as a prefix of every service; it is not asked at all.
func TestTasksOfNoServiceAreNone(t *testing.T) {
	m := &manager{} // no client: reaching Docker would panic
	for _, id := range []string{"", "  "} {
		resp, err := m.ServiceTaskList(context.Background(), id, nil)
		assert.NoError(t, err)
		if assert.NotNil(t, resp) {
			assert.Empty(t, resp.Items)
		}
	}
}

func runningTask(started time.Time) []swarm.Task {
	return []swarm.Task{{
		ID:     "task1",
		NodeID: "node1",
		Status: swarm.TaskStatus{State: swarm.TaskStateRunning, Timestamp: started},
	}}
}

// A container that has run long enough is the one, at the first look.
func TestRunningTaskOldEnoughIsFoundAtOnce(t *testing.T) {
	m, d := newDaemon(t, "/tasks", runningTask(time.Now().Add(-time.Hour)))
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 15*time.Second, 3, time.Second, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, task) {
		assert.Equal(t, "task1", task.ID)
	}
	assert.Equal(t, 1, d.calls)
}

// One just started is waited for until it has run long enough, though the
// retries the caller allows are spent before: a deployment right after another
// finds the container the first one started.
func TestRunningTaskJustStartedIsWaitedFor(t *testing.T) {
	m, _ := newDaemon(t, "/tasks", runningTask(time.Now().Add(-100*time.Millisecond)))
	start := time.Now()
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 400*time.Millisecond, 0, 0, nil)
	assert.NoError(t, err)
	assert.NotNil(t, task, "found once it has run 400ms")
	assert.Less(t, time.Since(start), time.Second, "waited for it, not a retry delay")
}

// With none running, it looks the times asked, the delay apart, and no longer:
// the time a container needs to grow up is not waited for one there is not.
func TestRunningTaskNoneLooksAsManyTimesAsAsked(t *testing.T) {
	m, d := newDaemon(t, "/tasks", []swarm.Task{})
	start := time.Now()
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 10*time.Second, 2, 50*time.Millisecond, nil)
	assert.NoError(t, err)
	assert.Nil(t, task)
	assert.Equal(t, 3, d.calls, "the first look and two more - the last after the last wait")
	assert.Less(t, time.Since(start), time.Second)
}

// A container that never grows up enough - one restarting over and over, or
// stamped ahead of this clock - is waited for minRunningDuration past the
// retries, no longer.
func TestRunningTaskWaitIsBounded(t *testing.T) {
	m, _ := newDaemon(t, "/tasks", runningTask(time.Now().Add(time.Hour)))
	start := time.Now()
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 200*time.Millisecond, 1, 50*time.Millisecond, nil)
	assert.NoError(t, err)
	assert.Nil(t, task)
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 250*time.Millisecond)
	assert.Less(t, elapsed, time.Second)
}

// Canceled, it stops waiting: a deployment canceled while it looks.
func TestRunningTaskWaitEndsWhenCanceled(t *testing.T) {
	m, _ := newDaemon(t, "/tasks", []swarm.Task{})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, _, err := m.ServiceTaskGetRunning(ctx, "svc", 10*time.Second, 100, time.Second, nil)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), time.Second)
}

func taskOf(id, image string, desired swarm.TaskState, started time.Time) swarm.Task {
	return swarm.Task{
		ID:           id,
		NodeID:       "node1",
		Spec:         swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}},
		DesiredState: desired,
		Status:       swarm.TaskStatus{State: swarm.TaskStateRunning, Timestamp: started},
	}
}

func serviceOf(image string) *swarm.Service {
	return &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Image: image},
	}}}
}

// Of the service's current spec, the container a clone's placeholder ran in is
// not taken, though it has run longer: the clone's own is, once it has run long
// enough. The digest swarm pins an image to changes nothing.
func TestRunningTaskOfTheCurrentSpecOnly(t *testing.T) {
	m, _ := newDaemon(t, "/tasks", []swarm.Task{
		taskOf("placeholder", "hivepaas/placeholder:1", swarm.TaskStateRunning, time.Now().Add(-time.Hour)),
		taskOf("own", "busybox:1.37@sha256:abc", swarm.TaskStateRunning, time.Now().Add(-time.Minute)),
	})
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 15*time.Second, 0, 0, nil,
		OfCurrentSpec(serviceOf("busybox:1.37")))
	assert.NoError(t, err)
	if assert.NotNil(t, task) {
		assert.Equal(t, "own", task.ID)
	}
}

// A container swarm is shutting down is not taken, of whatever spec.
func TestRunningTaskBeingShutDownIsNotTaken(t *testing.T) {
	m, _ := newDaemon(t, "/tasks", []swarm.Task{
		taskOf("going", "busybox:1.37", swarm.TaskStateShutdown, time.Now().Add(-time.Hour)),
	})
	task, _, err := m.ServiceTaskGetRunning(context.Background(), "svc", 15*time.Second, 0, 0, nil,
		OfCurrentSpec(serviceOf("busybox:1.37")))
	assert.NoError(t, err)
	assert.Nil(t, task)
}
