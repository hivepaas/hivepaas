package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type ConfigListOption func(*client.ConfigListOptions)

func (m *manager) ConfigList(
	ctx context.Context,
	options ...ConfigListOption,
) (*client.ConfigListResult, error) {
	opts := client.ConfigListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ConfigList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ConfigInspectOption func(options *client.ConfigInspectOptions)

func (m *manager) ConfigInspect(
	ctx context.Context,
	configID string,
	options ...ConfigInspectOption,
) (*client.ConfigInspectResult, error) {
	opts := client.ConfigInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ConfigInspect(ctx, configID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ConfigCreateOption func(*client.ConfigCreateOptions)

func (m *manager) ConfigCreate(
	ctx context.Context,
	name string,
	data []byte,
	options ...ConfigCreateOption,
) (*client.ConfigCreateResult, error) {
	opts := client.ConfigCreateOptions{}
	opts.Spec.Name = name
	opts.Spec.Data = data
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ConfigCreate(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ConfigRemoveOption func(options *client.ConfigRemoveOptions)

func (m *manager) ConfigRemove(
	ctx context.Context,
	configID string,
	options ...ConfigRemoveOption,
) (*client.ConfigRemoveResult, error) {
	opts := client.ConfigRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ConfigRemove(ctx, configID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
