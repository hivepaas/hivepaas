package docker

import (
	"context"
	"sync/atomic"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return nil, hperrors.NewInfra(err)
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

// NodeListByIDs lists the nodes of these ids, in one call, and only them: see
// byids.go. A node's hostname is its name to Docker, and not an id here.
func (m *manager) NodeListByIDs(
	ctx context.Context,
	nodeIDs []string,
	options ...NodeListOption,
) (*client.NodeListResult, error) {
	wanted := wantedIDs(nodeIDs)
	if len(wanted) == 0 {
		return &client.NodeListResult{Items: []swarm.Node{}}, nil
	}
	options = append(options, func(opts *client.NodeListOptions) {
		for id := range wanted {
			FilterAdd(&opts.Filters, "id", id)
		}
	})
	resp, err := m.NodeList(ctx, options...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Items = keepWanted(resp.Items, wanted, func(n *swarm.Node) string { return n.ID })
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
		return nil, hperrors.NewInfra(err)
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
			return nil, hperrors.Wrap(err)
		}
		version = &resp.Node.Version
	}
	opts.Version = *version

	resp, err := m.client.NodeUpdate(ctx, nodeID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
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
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

var (
	currentNodeID string
)

// NodeCurrentID is the swarm node id of the daemon this manager talks to.
//
// An empty id is an error here rather than a value. The daemon answers with one
// when it is not part of a swarm, and every caller is asking this in order to
// place something - a directory to create, a build to run, a volume to pin - so
// an empty answer does not mean "unknown", it means the work gets placed nowhere
// in particular. That is how a volume asked to pin to the current node quietly
// became a volume pinned to none.
//
// Not cached until it is real, so a daemon that joins a swarm later starts
// answering without a restart.
func (m *manager) NodeCurrentID(ctx context.Context) (string, error) {
	if currentNodeID != "" {
		return currentNodeID, nil
	}
	resp, err := m.SystemInfo(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if resp.Info.Swarm.NodeID == "" {
		return "", hperrors.NewNotFound("Swarm node").
			WithMsgLog("docker daemon is not part of a swarm")
	}
	currentNodeID = resp.Info.Swarm.NodeID
	return currentNodeID, nil
}

// currentSwarmClusterID is the cluster ID once a daemon has answered with one.
// Atomic, because license checks ask for it from requests running at once.
var currentSwarmClusterID atomic.Pointer[string]

// SwarmClusterID is the ID of the swarm the daemon this manager talks to is part
// of. A license key names the installation it is for by this ID.
//
// A daemon that is a swarm manager reports it; a worker, or a daemon in no
// swarm, does not, and that is NotFound rather than an empty ID: an empty ID
// would match no key while looking like an answer. HivePaaS runs on managers.
//
// Not cached until it is real, as NodeCurrentID; after that it is, because a
// swarm keeps its ID for as long as it exists.
func (m *manager) SwarmClusterID(ctx context.Context) (string, error) {
	if id := currentSwarmClusterID.Load(); id != nil {
		return *id, nil
	}
	resp, err := m.SystemInfo(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	cluster := resp.Info.Swarm.Cluster
	if cluster == nil || cluster.ID == "" {
		return "", hperrors.NewNotFound("Swarm cluster").
			WithMsgLog("docker daemon reports no swarm cluster: it is in no swarm, or not a manager of one")
	}
	id := cluster.ID
	currentSwarmClusterID.Store(&id)
	return id, nil
}
