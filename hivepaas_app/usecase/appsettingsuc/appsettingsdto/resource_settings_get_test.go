package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// Swap and swappiness are read each on its own: either may be set without the
// other.
func TestTransformMemoryReadsSwapAndSwappinessApart(t *testing.T) {
	swappiness := int64(10)
	resp := TransformMemory(&swarm.TaskSpec{Resources: &swarm.ResourceRequirements{MemorySwappiness: &swappiness}})
	if assert.NotNil(t, resp.Swappiness, "swappiness alone") {
		assert.Equal(t, int64(10), *resp.Swappiness)
	}
	assert.Nil(t, resp.Swap)

	swap := int64(64 * unit.MB)
	resp = TransformMemory(&swarm.TaskSpec{Resources: &swarm.ResourceRequirements{SwapBytes: &swap}})
	if assert.NotNil(t, resp.Swap, "swap alone") {
		assert.Equal(t, unit.DataSize(swap), *resp.Swap)
	}
	assert.Nil(t, resp.Swappiness)
}
