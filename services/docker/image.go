package docker

import (
	"context"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type ImageListOption func(*client.ImageListOptions)

func (m *manager) ImageList(
	ctx context.Context,
	options ...ImageListOption,
) (*client.ImageListResult, error) {
	opts := client.ImageListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImageList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ImagePullOption func(*client.ImagePullOptions)

func (m *manager) ImagePull(
	ctx context.Context,
	name string,
	options ...ImagePullOption,
) (client.ImagePullResponse, error) {
	opts := client.ImagePullOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImagePull(ctx, name, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

type ImagePushOption func(*client.ImagePushOptions)

func (m *manager) ImagePush(
	ctx context.Context,
	name string,
	options ...ImagePushOption,
) (client.ImagePushResponse, error) {
	opts := client.ImagePushOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImagePush(ctx, name, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

type ImageRemoveOption func(*client.ImageRemoveOptions)

func (m *manager) ImageRemove(
	ctx context.Context,
	imageID string,
	options ...ImageRemoveOption,
) (*client.ImageRemoveResult, error) {
	opts := client.ImageRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImageRemove(ctx, imageID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ImageInspectOption client.ImageInspectOption

func (m *manager) ImageInspect(
	ctx context.Context,
	imageID string,
	options ...ImageInspectOption,
) (*client.ImageInspectResult, error) {
	opts := make([]client.ImageInspectOption, 0, len(options))
	for _, opt := range options {
		opts = append(opts, opt)
	}
	resp, err := m.client.ImageInspect(ctx, imageID, opts...)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ImagePruneOption func(options *client.ImagePruneOptions)

func (m *manager) ImagePrune(
	ctx context.Context,
	danglingOnly bool,
	onlyObjectsOlderThan time.Duration,
	options ...ImagePruneOption,
) (*client.ImagePruneResult, error) {
	opts := client.ImagePruneOptions{}
	if danglingOnly {
		FilterAdd(&opts.Filters, "dangling", "true")
	}
	if onlyObjectsOlderThan > 0 {
		FilterAdd(&opts.Filters, "until", onlyObjectsOlderThan.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ImagePrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
