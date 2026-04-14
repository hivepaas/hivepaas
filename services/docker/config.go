package docker

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type ConfigListOption func(*swarm.ConfigListOptions)

func (m *manager) ConfigList(ctx context.Context, options ...ConfigListOption) ([]swarm.Config, error) {
	opts := swarm.ConfigListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ConfigList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ConfigInspect(ctx context.Context, configID string) (*swarm.Config, error) {
	resp, _, err := m.client.ConfigInspectWithRaw(ctx, configID)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ConfigSpecOption func(*swarm.ConfigSpec)

func (m *manager) ConfigCreate(
	ctx context.Context,
	name string,
	data []byte,
	options ...ConfigSpecOption,
) (*swarm.ConfigCreateResponse, error) {
	spec := swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Name: name,
		},
		Data: data,
	}
	for _, opt := range options {
		opt(&spec)
	}
	resp, err := m.client.ConfigCreate(ctx, spec)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ConfigRemove(ctx context.Context, configID string) error {
	err := m.client.ConfigRemove(ctx, configID)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}
