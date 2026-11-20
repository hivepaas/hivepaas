package clusterserviceimpl

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// removalDockerManager is a swarm of one service being removed: its tasks are
// read from tasksAt, by how many times the tasks have been listed so far.
type removalDockerManager struct {
	docker.Manager

	mu        sync.Mutex
	missing   bool
	removed   []string
	taskLists int
	tasksAt   func(call int) []swarm.Task
	nodes     []swarm.Node
}

func (f *removalDockerManager) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	if f.missing {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.ServiceInspectResult{Service: swarm.Service{ID: serviceID}}, nil
}

func (f *removalDockerManager) ServiceRemove(
	_ context.Context, serviceID string, _ ...docker.ServiceRemoveOption,
) (*client.ServiceRemoveResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, serviceID)
	return &client.ServiceRemoveResult{}, nil
}

func (f *removalDockerManager) TaskList(
	_ context.Context, _ ...docker.TaskListOption,
) (*client.TaskListResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	call := f.taskLists
	f.taskLists++
	return &client.TaskListResult{Items: f.tasksAt(call)}, nil
}

func (f *removalDockerManager) NodeList(
	_ context.Context, _ ...docker.NodeListOption,
) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: f.nodes}, nil
}

func task(serviceID, nodeID string, state swarm.TaskState) swarm.Task {
	return swarm.Task{ServiceID: serviceID, NodeID: nodeID, Status: swarm.TaskStatus{State: state}}
}

func node(id string, state swarm.NodeState) swarm.Node {
	return swarm.Node{ID: id, Status: swarm.NodeStatus{State: state}}
}

func removalService(f *removalDockerManager) *service {
	return &service{dockerManager: f, taskPollInterval: time.Millisecond}
}

// The removal is answered at once and the containers stop after: the wait goes
// on until the service's task is past running, and only then does the caller
// move on to the networks and volumes the container was attached to.
func TestServiceRemoveAndWaitWaitsForTheContainersToStop(t *testing.T) {
	f := &removalDockerManager{
		nodes: []swarm.Node{node("n1", swarm.NodeStateReady)},
		tasksAt: func(call int) []swarm.Task {
			if call < 3 {
				return []swarm.Task{task("svc", "n1", swarm.TaskStateRunning)}
			}
			return []swarm.Task{task("svc", "n1", swarm.TaskStateShutdown)}
		},
	}

	require.NoError(t, removalService(f).ServiceRemoveAndWait(context.Background(), "svc", 0, 0))

	assert.Equal(t, []string{"svc"}, f.removed)
	assert.Equal(t, 4, f.taskLists)
}

// What is not the removed service's, or is on a node that cannot report back,
// is not waited for.
func TestServiceRemoveAndWaitSkipsOtherServicesAndNodesThatAreDown(t *testing.T) {
	f := &removalDockerManager{
		nodes: []swarm.Node{node("up", swarm.NodeStateReady), node("down", swarm.NodeStateDown)},
		tasksAt: func(int) []swarm.Task {
			return []swarm.Task{
				task("other", "up", swarm.TaskStateRunning),
				task("svc", "down", swarm.TaskStateRunning),
				task("svc", "up", swarm.TaskStateComplete),
				task("svc", "up", swarm.TaskStateRejected),
			}
		},
	}

	require.NoError(t, removalService(f).ServiceRemoveAndWait(context.Background(), "svc", 0, 0))

	assert.Equal(t, 1, f.taskLists)
}

// A task not yet given a node has no container yet, but it is the service's
// and may get one: it is waited for like one that is running.
func TestServiceRemoveAndWaitWaitsForATaskNotYetOnANode(t *testing.T) {
	f := &removalDockerManager{
		tasksAt: func(call int) []swarm.Task {
			if call == 0 {
				return []swarm.Task{task("svc", "", swarm.TaskStatePending)}
			}
			return nil
		},
	}

	require.NoError(t, removalService(f).ServiceRemoveAndWait(context.Background(), "svc", 0, 0))

	assert.Equal(t, 2, f.taskLists)
}

// A service already gone has nothing to remove and nothing to wait for.
func TestServiceRemoveAndWaitOnAServiceAlreadyGone(t *testing.T) {
	f := &removalDockerManager{missing: true, tasksAt: func(int) []swarm.Task { return nil }}

	require.NoError(t, removalService(f).ServiceRemoveAndWait(context.Background(), "svc", 0, 0))

	assert.Empty(t, f.removed)
	assert.Zero(t, f.taskLists)
}

// A container that never stops ends the wait at its bound, with an error the
// caller logs rather than fails on.
func TestWaitServiceTasksStoppedEndsAtItsTimeout(t *testing.T) {
	f := &removalDockerManager{
		nodes:   []swarm.Node{node("n1", swarm.NodeStateReady)},
		tasksAt: func(int) []swarm.Task { return []swarm.Task{task("svc", "n1", swarm.TaskStateRunning)} },
	}

	err := removalService(f).waitServiceTasksStopped(context.Background(), "svc", 20*time.Millisecond)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestTaskStopTimeout(t *testing.T) {
	withGrace := func(grace *time.Duration) *swarm.Service {
		return &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{StopGracePeriod: grace},
		}}}
	}
	grace := func(d time.Duration) *time.Duration { return &d }

	assert.Equal(t, defaultStopGracePeriod+taskStopMargin, taskStopTimeout(&swarm.Service{}))
	assert.Equal(t, defaultStopGracePeriod+taskStopMargin, taskStopTimeout(withGrace(nil)))
	assert.Equal(t, time.Minute+taskStopMargin, taskStopTimeout(withGrace(grace(time.Minute))))
	assert.Equal(t, maxTaskStopWait, taskStopTimeout(withGrace(grace(time.Hour))))
}
