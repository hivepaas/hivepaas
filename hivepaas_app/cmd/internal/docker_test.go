package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

// An agent runs on every node, workers included, where the swarm's state cannot
// be read: it leaves syncing nodes, networks and volumes to the app.
func TestAnAgentDoesNotSyncSwarmObjectsAtStartup(t *testing.T) {
	assert.False(t, syncsSwarmObjectsAtStartup(config.RunModeAgent))

	assert.True(t, syncsSwarmObjectsAtStartup(config.RunModeApp))
	assert.True(t, syncsSwarmObjectsAtStartup(config.RunModeWorker))
	assert.True(t, syncsSwarmObjectsAtStartup(config.RunModeAppAndWorker))
}
