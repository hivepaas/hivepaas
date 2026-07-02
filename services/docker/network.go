package docker

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

const (
	NetworkDriverOverlay = "overlay"
	NetworkDriverBridge  = "bridge"
)

const (
	NetworkScopeSwarm = "swarm"
	NetworkScopeLocal = "local"
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
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NetworkListByIDs(
	ctx context.Context,
	networkIDOrNames []string,
	options ...NetworkListOption,
) (*client.NetworkListResult, error) {
	resp := &client.NetworkListResult{}
	if len(networkIDOrNames) == 0 {
		return resp, nil
	}

	if len(networkIDOrNames) == 1 {
		inspect, err := m.NetworkInspect(ctx, networkIDOrNames[0])
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.New(err)
		}
		if inspect != nil {
			resp.Items = append(resp.Items, network.Summary{Network: inspect.Network.Network})
		}
		return resp, nil
	}

	listResp, err := m.NetworkList(ctx, options...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	for i := range listResp.Items {
		net := &listResp.Items[i]
		if gofn.Contain(networkIDOrNames, net.Name) || gofn.Contain(networkIDOrNames, net.ID) {
			resp.Items = append(resp.Items, *net)
			continue
		}
	}

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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
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
	onlyObjectsOlderThan time.Duration,
	options ...NetworkPruneOption,
) (*client.NetworkPruneResult, error) {
	opts := client.NetworkPruneOptions{}
	if onlyObjectsOlderThan > 0 {
		FilterAdd(&opts.Filters, "until", onlyObjectsOlderThan.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NetworkPrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
