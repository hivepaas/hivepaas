package clusterserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

const (
	// defaultStopGracePeriod is how long docker gives a container to stop when its
	// service does not say: SIGTERM, then this long, then SIGKILL.
	defaultStopGracePeriod = 10 * time.Second

	// taskStopMargin is what the wait allows beyond the grace period: for the
	// node to hear of the removal, kill the container and report back.
	taskStopMargin = 30 * time.Second

	// maxTaskStopWait bounds the wait whatever grace period an app asks for, so
	// that removing an app that takes an hour to stop does not hold its deletion
	// - and the database transaction around it - for an hour.
	maxTaskStopWait = 5 * time.Minute

	defaultTaskPollInterval = time.Second
)

// ServiceRemoveAndWait removes a service, then waits for its containers to stop.
//
// Docker answers a service removal at once and stops the containers after, and
// swarm lets the service's networks go as soon as no service names them - while
// the node still has the old containers attached. What happens next is the
// reason for the wait: swarm hands the freed subnet to the next network created,
// the lowest free one, and a task of that network on the node that still has the
// old one is rejected with "Pool overlaps with other one on this address space".
// An app deleted with its project, then a project created, left the new
// project's apps unable to start until the node had cleaned up.
//
// So what comes after a removal - the app's networks, its volumes, its project's
// networks - waits for this. The wait is not fatal: a node that does not report
// back is no reason to leave an app half deleted, and the cost of going on is
// only the race above.
func (s *service) ServiceRemoveAndWait(
	ctx context.Context,
	serviceID string,
	retryMax int,
	retryDelay time.Duration,
) error {
	if serviceID == "" {
		return nil
	}

	// Read before it goes: afterwards nothing says how long its containers take.
	timeout := defaultStopGracePeriod + taskStopMargin
	resp, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	switch {
	case err == nil:
		timeout = taskStopTimeout(&resp.Service)
	case errors.Is(err, hperrors.ErrNotFound):
		return nil
	}

	if err := s.ServiceRemove(ctx, serviceID, retryMax, retryDelay); err != nil {
		return hperrors.Wrap(err)
	}

	if err := s.waitServiceTasksStopped(ctx, serviceID, timeout); err != nil {
		logging.Warnf("cluster: the containers of removed service %s did not all stop in %s: %v",
			serviceID, timeout, err)
	}
	return nil
}

// taskStopTimeout is how long the containers of a service may take to stop once
// it is removed: its grace period, and the margin.
func taskStopTimeout(svc *swarm.Service) time.Duration {
	grace := defaultStopGracePeriod
	if spec := svc.Spec.TaskTemplate.ContainerSpec; spec != nil && spec.StopGracePeriod != nil {
		grace = *spec.StopGracePeriod
	}
	return min(grace+taskStopMargin, maxTaskStopWait)
}

// waitServiceTasksStopped waits until no task of a removed service holds a
// container on a node that can still report it.
//
// The tasks are read without a service filter: docker resolves that filter to
// a service first, and a removed service is not found. A task on a node that is
// not ready is not waited for - it will not report its container stopped until
// the node comes back, and the node removes it then.
func (s *service) waitServiceTasksStopped(ctx context.Context, serviceID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	interval := s.taskPollInterval
	if interval <= 0 {
		interval = defaultTaskPollInterval
	}

	for {
		live, err := s.serviceLiveTasks(ctx, serviceID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if live == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return hperrors.Wrap(ctx.Err()).WithMsgLog("%d task(s) still running", live)
		case <-time.After(interval):
		}
	}
}

// serviceLiveTasks counts the tasks of the service that may still have a
// container: not yet past running, on a node that is ready.
func (s *service) serviceLiveTasks(ctx context.Context, serviceID string) (int, error) {
	tasks, err := s.dockerManager.TaskList(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}

	var nodes map[string]bool // node id -> ready; read only when a task needs it
	live := 0
	for i := range tasks.Items {
		task := &tasks.Items[i]
		if task.ServiceID != serviceID || !taskMayHoldContainer(task.Status.State) {
			continue
		}
		if task.NodeID != "" {
			if nodes == nil {
				if nodes, err = s.readyNodes(ctx); err != nil {
					return 0, hperrors.Wrap(err)
				}
			}
			if !nodes[task.NodeID] {
				continue
			}
		}
		live++
	}
	return live, nil
}

func (s *service) readyNodes(ctx context.Context) (map[string]bool, error) {
	resp, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ready := make(map[string]bool, len(resp.Items))
	for i := range resp.Items {
		ready[resp.Items[i].ID] = resp.Items[i].Status.State == swarm.NodeStateReady
	}
	return ready, nil
}

// taskMayHoldContainer reports whether a task in this state may still have a
// container: every state up to running. The ones after it - complete, shutdown,
// failed, rejected, orphaned, remove - are of a container that has stopped or
// never was.
func taskMayHoldContainer(state swarm.TaskState) bool {
	switch state {
	case swarm.TaskStateNew, swarm.TaskStateAllocated, swarm.TaskStatePending, swarm.TaskStateAssigned,
		swarm.TaskStateAccepted, swarm.TaskStatePreparing, swarm.TaskStateReady, swarm.TaskStateStarting,
		swarm.TaskStateRunning:
		return true
	case swarm.TaskStateComplete, swarm.TaskStateShutdown, swarm.TaskStateFailed, swarm.TaskStateRejected,
		swarm.TaskStateRemove, swarm.TaskStateOrphaned:
		return false
	}
	return false
}
