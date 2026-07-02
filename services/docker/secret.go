package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type SecretListOption func(*client.SecretListOptions)

func (m *manager) SecretList(
	ctx context.Context,
	options ...SecretListOption,
) (*client.SecretListResult, error) {
	opts := client.SecretListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SecretList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type SecretInspectOption func(options *client.SecretInspectOptions)

func (m *manager) SecretInspect(
	ctx context.Context,
	configID string,
	options ...SecretInspectOption,
) (*client.SecretInspectResult, error) {
	opts := client.SecretInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SecretInspect(ctx, configID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type SecretCreateOption func(*client.SecretCreateOptions)

func (m *manager) SecretCreate(
	ctx context.Context,
	name string,
	data []byte,
	options ...SecretCreateOption,
) (*client.SecretCreateResult, error) {
	opts := client.SecretCreateOptions{}
	opts.Spec.Name = name
	opts.Spec.Data = data
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SecretCreate(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type SecretRemoveOption func(options *client.SecretRemoveOptions)

func (m *manager) SecretRemove(
	ctx context.Context,
	configID string,
	options ...SecretRemoveOption,
) (*client.SecretRemoveResult, error) {
	opts := client.SecretRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.SecretRemove(ctx, configID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
