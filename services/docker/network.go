package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

const (
	NetworkDriverOverlay = "overlay"
	NetworkDriverBridge  = "bridge"
)

const (
	NetworkScopeSwarm = "swarm"
	NetworkScopeLocal = "local"
)

type NetworkListOption func(*network.ListOptions)

func (m *manager) NetworkList(
	ctx context.Context,
	options ...NetworkListOption,
) ([]network.Summary, error) {
	opts := network.ListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

type NetworkCreateOption func(*network.CreateOptions)

func (m *manager) NetworkCreate(
	ctx context.Context,
	name string,
	options ...NetworkCreateOption,
) (*network.CreateResponse, error) {
	opts := network.CreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkCreate(ctx, name, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NetworkRemove(ctx context.Context, idOrName string) error {
	err := m.client.NetworkRemove(ctx, idOrName)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

type NetworkInspectOption func(*network.InspectOptions)

func (m *manager) NetworkInspect(
	ctx context.Context,
	name string,
	options ...NetworkInspectOption,
) (*network.Inspect, error) {
	opts := network.InspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkInspect(ctx, name, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NetworkExists(ctx context.Context, name string) bool {
	_, err := m.NetworkInspect(ctx, name)
	return err == nil
}

type NetworksPruneOption func(options *filters.Args)

func (m *manager) NetworksPrune(
	ctx context.Context,
	onlyObjectsOlderThan time.Duration,
	options ...NetworksPruneOption,
) (*network.PruneReport, error) {
	opts := filters.Args{}
	if onlyObjectsOlderThan > 0 {
		FilterAdd(&opts, "until", onlyObjectsOlderThan.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworksPrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
