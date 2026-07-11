package docker

import (
	"context"

	"github.com/moby/moby/client"
)

type EventsOption func(options *client.EventsListOptions)

func (m *manager) Events(
	ctx context.Context,
	options ...EventsOption,
) *client.EventsResult {
	opts := client.EventsListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp := m.client.Events(ctx, opts)
	return &resp
}
