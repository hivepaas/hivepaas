package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type SwarmInspectOption func(options *client.SwarmInspectOptions)

func (m *manager) SwarmInspect(
	ctx context.Context,
	options ...SwarmInspectOption,
) (*client.SwarmInspectResult, error) {
	opts := client.SwarmInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SwarmInspect(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
