package docker

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type SecretListOption func(*swarm.SecretListOptions)

func (m *manager) SecretList(ctx context.Context, options ...SecretListOption) ([]swarm.Secret, error) {
	opts := swarm.SecretListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SecretList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) SecretInspect(ctx context.Context, secretID string) (*swarm.Secret, error) {
	resp, _, err := m.client.SecretInspectWithRaw(ctx, secretID)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type SecretSpecOption func(*swarm.SecretSpec)

func (m *manager) SecretCreate(
	ctx context.Context,
	name string,
	data []byte,
	options ...SecretSpecOption,
) (*swarm.SecretCreateResponse, error) {
	spec := swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name: name,
		},
		Data: data,
	}
	for _, opt := range options {
		opt(&spec)
	}
	resp, err := m.client.SecretCreate(ctx, spec)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) SecretRemove(ctx context.Context, secretID string) error {
	err := m.client.SecretRemove(ctx, secretID)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}
