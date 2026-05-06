package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
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
	desiredState string, // running | shutdown | accepted
	options ...TaskListOption,
) (*client.TaskListResult, error) {
	options = append(options, func(opts *client.TaskListOptions) {
		FilterAdd(&opts.Filters, "service", serviceID)
		if desiredState != "" {
			FilterAdd(&opts.Filters, "desired-state", desiredState)
		}
	})
	return m.TaskList(ctx, options...)
}
