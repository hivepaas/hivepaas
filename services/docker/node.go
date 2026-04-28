package docker

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type NodeStatus string

const (
	NodeStatusUnknown      = NodeStatus(swarm.NodeStateUnknown)
	NodeStatusDown         = NodeStatus(swarm.NodeStateDown)
	NodeStatusReady        = NodeStatus(swarm.NodeStateReady)
	NodeStatusDisconnected = NodeStatus(swarm.NodeStateDisconnected)
)

var (
	AllNodeStatuses = []NodeStatus{NodeStatusUnknown, NodeStatusDown, NodeStatusReady, NodeStatusDisconnected}
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

type NodeListOption func(*swarm.NodeListOptions)

func (m *manager) NodeList(
	ctx context.Context,
	options ...NodeListOption,
) ([]swarm.Node, error) {
	opts := swarm.NodeListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.NodeList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) NodeManagerList(
	ctx context.Context,
	options ...NodeListOption,
) ([]swarm.Node, error) {
	options = append(options, func(opts *swarm.NodeListOptions) {
		FilterAdd(&opts.Filters, "role", "manager")
	})
	return m.NodeList(ctx, options...)
}

func (m *manager) NodeInspect(
	ctx context.Context,
	nodeID string,
) (*swarm.Node, []byte, error) {
	resp, data, err := m.client.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		return nil, nil, apperrors.NewInfra(err)
	}
	return &resp, data, nil
}

func (m *manager) NodeUpdate(
	ctx context.Context,
	nodeID string,
	version *swarm.Version,
	spec *swarm.NodeSpec,
) error {
	if spec == nil {
		return nil
	}
	if version == nil {
		resp, _, err := m.client.NodeInspectWithRaw(ctx, nodeID)
		if err != nil {
			return apperrors.NewInfra(err)
		}
		version = &resp.Version
	}
	err := m.client.NodeUpdate(ctx, nodeID, *version, *spec)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

type NodeRemoveOption func(*swarm.NodeRemoveOptions)

func NodeRemoveForce(force bool) NodeRemoveOption {
	return func(opts *swarm.NodeRemoveOptions) {
		opts.Force = force
	}
}

func (m *manager) NodeRemove(
	ctx context.Context,
	nodeID string,
	options ...NodeRemoveOption,
) error {
	opts := swarm.NodeRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	err := m.client.NodeRemove(ctx, nodeID, opts)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}
