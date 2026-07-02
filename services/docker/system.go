package docker

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type SystemInfoOption func(options *client.InfoOptions)

func (m *manager) SystemInfo(
	ctx context.Context,
	options ...SystemInfoOption,
) (*client.SystemInfoResult, error) {
	opts := client.InfoOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.Info(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
