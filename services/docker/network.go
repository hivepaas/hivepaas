package docker

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	NetworkDriverOverlay = "overlay"
	NetworkDriverBridge  = "bridge"
)

const (
	NetworkScopeSwarm = "swarm"
	NetworkScopeLocal = "local"
)

const (
	NetworkOptionDriverMTU   = "com.docker.network.driver.mtu"
	DefaultOverlayNetworkMTU = "1380"
)

type NetworkListOption func(*client.NetworkListOptions)

func (m *manager) NetworkList(
	ctx context.Context,
	options ...NetworkListOption,
) (*client.NetworkListResult, error) {
	opts := client.NetworkListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkList(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

// NetworkListByIDs lists the networks of these ids, in one call, and only them:
// see byids.go.
func (m *manager) NetworkListByIDs(
	ctx context.Context,
	networkIDs []string,
	options ...NetworkListOption,
) (*client.NetworkListResult, error) {
	wanted := wantedIDs(networkIDs)
	if len(wanted) == 0 {
		return &client.NetworkListResult{Items: []network.Summary{}}, nil
	}
	options = append(options, func(opts *client.NetworkListOptions) {
		for id := range wanted {
			FilterAdd(&opts.Filters, "id", id)
		}
	})
	resp, err := m.NetworkList(ctx, options...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Items = keepWanted(resp.Items, wanted, func(n *network.Summary) string { return n.ID })
	return resp, nil
}

type NetworkCreateOption func(*client.NetworkCreateOptions)

func (m *manager) NetworkCreate(
	ctx context.Context,
	name string,
	options ...NetworkCreateOption,
) (*client.NetworkCreateResult, error) {
	opts := client.NetworkCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkCreate(ctx, name, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type NetworkRemoveOption func(*client.NetworkRemoveOptions)

func (m *manager) NetworkRemove(
	ctx context.Context,
	idOrName string,
	options ...NetworkRemoveOption,
) (*client.NetworkRemoveResult, error) {
	opts := client.NetworkRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkRemove(ctx, idOrName, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type NetworkInspectOption func(*client.NetworkInspectOptions)

func (m *manager) NetworkInspect(
	ctx context.Context,
	name string,
	options ...NetworkInspectOption,
) (*client.NetworkInspectResult, error) {
	opts := client.NetworkInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkInspect(ctx, name, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NetworkExists(ctx context.Context, name string) bool {
	_, err := m.NetworkInspect(ctx, name)
	return err == nil
}

type NetworkPruneOption func(*client.NetworkPruneOptions)

func (m *manager) NetworkPrune(
	ctx context.Context,
	generalRetention time.Duration,
	options ...NetworkPruneOption,
) (*client.NetworkPruneResult, error) {
	opts := client.NetworkPruneOptions{}
	if generalRetention > 0 {
		FilterAdd(&opts.Filters, "until", generalRetention.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkPrune(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}
