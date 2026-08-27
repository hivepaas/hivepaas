package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RegistryLoginOption func(*client.RegistryLoginOptions)

func (m *manager) RegistryLogin(
	ctx context.Context,
	options ...RegistryLoginOption,
) (*client.RegistryLoginResult, error) {
	opts := client.RegistryLoginOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.RegistryLogin(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}
