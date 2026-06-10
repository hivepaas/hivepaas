package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type TaskState string

const (
	TaskStateRunning  TaskState = "running"
	TaskStateShutdown TaskState = "shutdown"
	TaskStateAccepted TaskState = "accepted"
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
	desiredState TaskState,
	options ...TaskListOption,
) (*client.TaskListResult, error) {
	options = append(options, func(opts *client.TaskListOptions) {
		FilterAdd(&opts.Filters, "service", serviceID)
		if desiredState != "" {
			FilterAdd(&opts.Filters, "desired-state", string(desiredState))
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
