package docker

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceTaskList(
	ctx context.Context,
	serviceID string,
	desiredStates []swarm.TaskState,
	options ...TaskListOption,
) (*client.TaskListResult, error) {
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ServiceTaskGetRunning(
	ctx context.Context,
	serviceID string,
	minRunningDuration time.Duration,
	maxRetry int,
	retryDelay time.Duration,
	ignoreNodeIDs []string,
) (running *swarm.Task, all *client.TaskListResult, err error) {
	return m.serviceTaskGetRunning(ctx, serviceID, minRunningDuration, -1,
		maxRetry, retryDelay, ignoreNodeIDs)
}

func (m *manager) serviceTaskGetRunning(
	ctx context.Context,
	serviceID string,
	minRunningDuration time.Duration,
	retry int,
	maxRetry int,
	retryDelay time.Duration,
	ignoreNodeIDs []string,
) (running *swarm.Task, all *client.TaskListResult, err error) {
	if retry >= maxRetry {
		return nil, nil, nil
	}
	listResp, err := m.ServiceTaskList(ctx, serviceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	timeNow := time.Now()
	waitDuration := retryDelay
	for i := range listResp.Items {
		t := &listResp.Items[i]
		if t.Status.State != swarm.TaskStateRunning || gofn.Contain(ignoreNodeIDs, t.NodeID) {
			continue
		}
		duration := timeNow.Sub(t.Status.Timestamp)
		if duration > minRunningDuration {
			return t, listResp, nil
		}
		waitDuration = min(waitDuration, minRunningDuration-duration)
	}

	time.Sleep(max(waitDuration, time.Second))
	return m.serviceTaskGetRunning(ctx, serviceID, minRunningDuration, retry+1,
		maxRetry, retryDelay, ignoreNodeIDs)
}
