package clustercleanupserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// swarmNode is a docker manager for a node of a swarm, a manager or not, whose
// swarm has no service.
type swarmNode struct {
	docker.Manager
	manager      bool
	servicesSeen bool
}

func (n *swarmNode) SystemInfo(context.Context, ...docker.SystemInfoOption) (*client.SystemInfoResult, error) {
	res := &client.SystemInfoResult{}
	res.Info.Swarm = swarm.Info{NodeID: "n1", ControlAvailable: n.manager}
	return res, nil
}

func (n *swarmNode) ServiceList(context.Context, ...docker.ServiceListOption) (*client.ServiceListResult, error) {
	n.servicesSeen = true
	return &client.ServiceListResult{}, nil
}

func cleanupDataFor() *clusterCleanupData {
	return &clusterCleanupData{Output: &entity.ClusterNodeCleanupOutput{}}
}

// The cleanup runs on every node, through its agent. Listing services on a
// worker is refused by docker - "This node is not a swarm manager" - and failed
// the whole system cleanup job.
func TestTempServicesAreLeftToTheManager(t *testing.T) {
	worker := &swarmNode{}
	assert.NoError(t, (&service{dockerManager: worker}).cleanupTempServices(context.Background(), cleanupDataFor()))
	assert.False(t, worker.servicesSeen, "a worker does not ask for the swarm's services")

	manager := &swarmNode{manager: true}
	assert.NoError(t, (&service{dockerManager: manager}).cleanupTempServices(context.Background(), cleanupDataFor()))
	assert.True(t, manager.servicesSeen)
}
