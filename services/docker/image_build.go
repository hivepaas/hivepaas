package docker

import (
	"context"
	"io"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type ImageBuildOption func(options *client.ImageBuildOptions)

func (m *manager) ImageBuild(
	ctx context.Context,
	buildContext io.Reader,
	options ...ImageBuildOption,
) (*client.ImageBuildResult, error) {
	opts := client.ImageBuildOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImageBuild(ctx, buildContext, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ImageBuildCancelOption func(options *client.BuildCancelOptions)

func (m *manager) ImageBuildCancel(
	ctx context.Context,
	buildID string,
	options ...ImageBuildCancelOption,
) (*client.BuildCancelResult, error) {
	opts := client.BuildCancelOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.BuildCancel(ctx, buildID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type BuildCachePruneOption func(options *client.BuildCachePruneOptions)

func (m *manager) BuildCachePrune(
	ctx context.Context,
	options ...BuildCachePruneOption,
) (*client.BuildCachePruneResult, error) {
	opts := client.BuildCachePruneOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.BuildCachePrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
