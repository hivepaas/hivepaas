package docker

import (
	"context"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type TaskListOption func(*client.TaskListOptions)

func (m *manager) TaskList(
	ctx context.Context,
	options ...TaskListOption,
) (*client.TaskListResult, error) {
	opts := client.TaskListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.TaskList(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceTaskList(
	ctx context.Context,
	serviceID string,
	desiredStates []swarm.TaskState,
	options ...TaskListOption,
) (*client.TaskListResult, error) {
	// No service, no tasks. Docker reads the filter as a prefix of a service's
	// name or id, and an empty one is a prefix of all: asked about an app never
	// deployed, it refused with "service is ambiguous" - or, on a swarm of one
	// service, answered that service's tasks as the app's.
	if strings.TrimSpace(serviceID) == "" {
		return &client.TaskListResult{Items: []swarm.Task{}}, nil
	}
	options = append(options, func(opts *client.TaskListOptions) {
		FilterAdd(&opts.Filters, "service", serviceID)
		for _, state := range desiredStates {
			FilterAdd(&opts.Filters, "desired-state", string(state))
		}
	})
	return m.TaskList(ctx, options...)
}

type TaskInspectOption func(options *client.TaskInspectOptions)

func (m *manager) TaskInspect(
	ctx context.Context,
	taskID string,
	options ...TaskInspectOption,
) (*client.TaskInspectResult, error) {
	opts := client.TaskInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.TaskInspect(ctx, taskID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type TaskLogsOption func(*client.TaskLogsOptions)

func (m *manager) TaskLogs(
	ctx context.Context,
	containerID string,
	options ...TaskLogsOption,
) (client.TaskLogsResult, error) {
	if containerID == "" {
		return nil, nil
	}

	opts := client.TaskLogsOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.TaskLogs(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return resp, nil
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// minTaskLookInterval keeps the looks for a running task apart however short
// the wait asked for.
const minTaskLookInterval = 10 * time.Millisecond

// ServiceTaskGetRunning is a running task of the service that has run longer
// than minRunningDuration: one that has not may still fail to start. It looks
// maxRetry more times, retryDelay apart, for one; and a task that runs, not that
// long yet, is waited for until it has - minRunningDuration more at most - a
// container started as the caller came being what it is there for. It answers
// nil when there is none by then.
func (m *manager) ServiceTaskGetRunning(
	ctx context.Context,
	serviceID string,
	minRunningDuration time.Duration,
	maxRetry int,
	retryDelay time.Duration,
	ignoreNodeIDs []string,
) (running *swarm.Task, all *client.TaskListResult, err error) {
	start := time.Now()
	deadline := start.Add(time.Duration(max(maxRetry, 0)) * retryDelay)
	limit := deadline.Add(minRunningDuration)
	for {
		listResp, err := m.ServiceTaskList(ctx, serviceID, []swarm.TaskState{swarm.TaskStateRunning})
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}

		now := time.Now()
		next := now.Add(retryDelay)
		for i := range listResp.Items {
			t := &listResp.Items[i]
			if t.Status.State != swarm.TaskStateRunning || gofn.Contain(ignoreNodeIDs, t.NodeID) {
				continue
			}
			grown := t.Status.Timestamp.Add(minRunningDuration)
			if now.After(grown) {
				return t, listResp, nil
			}
			// Running, not long enough yet: looked at again once it has.
			if grown.After(deadline) {
				deadline = earliest(grown, limit)
			}
			if grown.Before(next) {
				next = grown
			}
		}

		if !now.Before(deadline) {
			return nil, nil, nil
		}
		wait := max(earliest(next, deadline).Sub(now), minTaskLookInterval)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, nil, hperrors.Wrap(ctx.Err())
		case <-timer.C:
		}
	}
}
