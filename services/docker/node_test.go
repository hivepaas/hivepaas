package docker

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/system"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// forgetSwarmClusterID drops the cached ID before and after a test, which
// otherwise would be answered by whichever test cached one first.
func forgetSwarmClusterID(t *testing.T) {
	t.Helper()
	currentSwarmClusterID.Store(nil)
	t.Cleanup(func() { currentSwarmClusterID.Store(nil) })
}

// The daemon is asked once: a swarm keeps its ID.
func TestSwarmClusterIDIsAskedOnce(t *testing.T) {
	forgetSwarmClusterID(t)
	m, d := newDaemon(t, "/info", system.Info{Swarm: swarm.Info{
		NodeID: "node1", Cluster: &swarm.ClusterInfo{ID: "cluster1"},
	}})
	for range 2 {
		id, err := m.SwarmClusterID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "cluster1", id)
	}
	assert.Equal(t, 1, d.calls)
}

// A worker, or a daemon in no swarm, reports no cluster. That is not an ID, and
// it is not kept: the daemon is asked again next time.
func TestNoSwarmClusterIsNotFound(t *testing.T) {
	for name, info := range map[string]swarm.Info{
		"in no swarm": {},
		"a worker":    {NodeID: "node1"},
		"an empty ID": {NodeID: "node1", Cluster: &swarm.ClusterInfo{}},
	} {
		t.Run(name, func(t *testing.T) {
			forgetSwarmClusterID(t)
			m, d := newDaemon(t, "/info", system.Info{Swarm: info})
			for range 2 {
				id, err := m.SwarmClusterID(context.Background())
				assert.ErrorIs(t, err, hperrors.ErrNotFound)
				assert.Empty(t, id)
			}
			assert.Equal(t, 2, d.calls)
		})
	}
}
