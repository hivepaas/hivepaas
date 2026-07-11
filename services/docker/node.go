package docker

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type NodeState string

const (
	NodeStatusUnknown      = NodeState(swarm.NodeStateUnknown)
	NodeStatusDown         = NodeState(swarm.NodeStateDown)
	NodeStatusReady        = NodeState(swarm.NodeStateReady)
	NodeStatusDisconnected = NodeState(swarm.NodeStateDisconnected)
)

var (
	AllNodeStates = []NodeState{NodeStatusUnknown, NodeStatusDown, NodeStatusReady, NodeStatusDisconnected}
)

type NodeRole string

const (
	NodeRoleManager = NodeRole(swarm.NodeRoleManager)
	NodeRoleWorker  = NodeRole(swarm.NodeRoleWorker)
)

var (
	AllNodeRoles = []NodeRole{NodeRoleManager, NodeRoleWorker}
)

type NodeAvailability string

const (
	NodeAvailabilityActive = NodeAvailability(swarm.NodeAvailabilityActive)
	NodeAvailabilityPause  = NodeAvailability(swarm.NodeAvailabilityPause)
	NodeAvailabilityDrain  = NodeAvailability(swarm.NodeAvailabilityDrain)
)

var (
	AllNodeAvailabilities = []NodeAvailability{NodeAvailabilityActive, NodeAvailabilityPause,
		NodeAvailabilityDrain}
)

type NodeListOption func(*client.NodeListOptions)

func (m *manager) NodeList(
	ctx context.Context,
	options ...NodeListOption,
) (*client.NodeListResult, error) {
	opts := client.NodeListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NodeList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NodeManagerList(
	ctx context.Context,
	options ...NodeListOption,
) (*client.NodeListResult, error) {
	options = append(options, func(opts *client.NodeListOptions) {
		FilterAdd(&opts.Filters, "role", "manager")
	})
	return m.NodeList(ctx, options...)
}

func (m *manager) NodeListByIDs(
	ctx context.Context,
	nodeIDOrNames []string,
	options ...NodeListOption,
) (*client.NodeListResult, error) {
	resp := &client.NodeListResult{}
	if len(nodeIDOrNames) == 0 {
		return resp, nil
	}

	if len(nodeIDOrNames) == 1 {
		inspect, err := m.NodeInspect(ctx, nodeIDOrNames[0])
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.New(err)
		}
		if inspect != nil {
			resp.Items = []swarm.Node{inspect.Node}
		}
		return resp, nil
	}

	listResp, err := m.NodeList(ctx, options...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	for i := range listResp.Items {
		node := &listResp.Items[i]
		if gofn.Contain(nodeIDOrNames, node.ID) || gofn.Contain(nodeIDOrNames, node.Spec.Name) {
			resp.Items = append(resp.Items, *node)
			continue
		}
	}

	return resp, nil
}

type NodeInspectOption func(*client.NodeInspectOptions)

func (m *manager) NodeInspect(
	ctx context.Context,
	nodeID string,
	options ...NodeInspectOption,
) (*client.NodeInspectResult, error) {
	opts := client.NodeInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NodeInspect(ctx, nodeID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) NodeUpdate(
	ctx context.Context,
	nodeID string,
	version *swarm.Version,
	spec *swarm.NodeSpec,
) (*client.NodeUpdateResult, error) {
	if spec == nil {
		return nil, nil
	}
	opts := client.NodeUpdateOptions{
		Spec: *spec,
	}

	if version == nil {
		resp, err := m.NodeInspect(ctx, nodeID)
		if err != nil {
			return nil, apperrors.New(err)
		}
		version = &resp.Node.Version
	}
	opts.Version = *version

	resp, err := m.client.NodeUpdate(ctx, nodeID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type NodeRemoveOption func(*client.NodeRemoveOptions)

func NodeRemoveForce(force bool) NodeRemoveOption {
	return func(opts *client.NodeRemoveOptions) {
		opts.Force = force
	}
}

func (m *manager) NodeRemove(
	ctx context.Context,
	nodeID string,
	options ...NodeRemoveOption,
) (*client.NodeRemoveResult, error) {
	opts := client.NodeRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NodeRemove(ctx, nodeID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

var (
	currentNodeID string
)

func (m *manager) NodeCurrentID(ctx context.Context) (string, error) {
	if currentNodeID != "" {
		return currentNodeID, nil
	}
	resp, err := m.SystemInfo(ctx)
	if err != nil {
		return "", apperrors.New(err)
	}
	currentNodeID = resp.Info.Swarm.NodeID
	return currentNodeID, nil
}
