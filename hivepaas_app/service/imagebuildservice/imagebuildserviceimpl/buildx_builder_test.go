package imagebuildserviceimpl

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

// The pattern has to match what the docker-container driver actually names its
// container. Getting this wrong is invisible: the filter returns nothing and the
// builder looks like it never started.
func Test_buildkitContainerPattern(t *testing.T) {
	pattern := buildkitContainerPattern("hivepaas_builder")
	re := regexp.MustCompile(pattern)

	// Docker stores container names with a leading slash; the CLI shows them without.
	assert.True(t, re.MatchString("/buildx_buildkit_hivepaas_builder0"))
	assert.True(t, re.MatchString("buildx_buildkit_hivepaas_builder0"))
	// A builder with several nodes has one container per node.
	assert.True(t, re.MatchString("/buildx_buildkit_hivepaas_builder1"))
	assert.True(t, re.MatchString("/buildx_buildkit_hivepaas_builder12"))

	assert.False(t, re.MatchString("/buildx_buildkit_other_builder0"))
	assert.False(t, re.MatchString("/buildx_buildkit_hivepaas_builder"), "the node index is required")
	assert.False(t, re.MatchString("/some_hivepaas_builder0"))
}

// Anchoring matters: without it a builder name that is a prefix of another would
// collect the other builder's containers and resize them.
func Test_buildkitContainerPattern_DoesNotMatchLongerNames(t *testing.T) {
	re := regexp.MustCompile(buildkitContainerPattern("web"))

	assert.True(t, re.MatchString("/buildx_buildkit_web0"))
	assert.False(t, re.MatchString("/buildx_buildkit_web-staging0"))
	assert.False(t, re.MatchString("/buildx_buildkit_myweb0"))
}

// The builder name comes from configuration, so a regex metacharacter in it must be
// matched literally instead of changing what the pattern means.
func Test_buildkitContainerPattern_QuotesTheBuilderName(t *testing.T) {
	re := regexp.MustCompile(buildkitContainerPattern("a.c"))

	assert.True(t, re.MatchString("/buildx_buildkit_a.c0"))
	assert.False(t, re.MatchString("/buildx_buildkit_abc0"))
}

// Docker refuses a memory update that does not carry a swap limit at least as large,
// and the buildkit container is created with neither. Sending Memory on its own is
// what produced "update the memoryswap at the same time" against a live daemon.
func Test_buildResourceUpdate_MemoryAlwaysCarriesSwap(t *testing.T) {
	update, skipped := buildResourceUpdate(&entity.ImageBuildResourceSettings{
		Mem: unit.DataSize(2 * 1024 * 1024 * 1024),
	})

	assert.Empty(t, skipped)
	assert.EqualValues(t, 2*1024*1024*1024, update.Memory)
	assert.EqualValues(t, 4*1024*1024*1024, update.MemorySwap,
		"an unset swap limit must default to what docker uses for `run -m X`")
	assert.GreaterOrEqual(t, update.MemorySwap, update.Memory)
}

func Test_buildResourceUpdate_ExplicitSwapIsKept(t *testing.T) {
	update, skipped := buildResourceUpdate(&entity.ImageBuildResourceSettings{
		Mem:     unit.DataSize(2 * 1024 * 1024 * 1024),
		MemSwap: unit.DataSize(3 * 1024 * 1024 * 1024),
	})

	assert.Empty(t, skipped)
	assert.EqualValues(t, 3*1024*1024*1024, update.MemorySwap)

	// memSwap equal to mem is how an operator asks for no swap; it must be honored.
	update, skipped = buildResourceUpdate(&entity.ImageBuildResourceSettings{
		Mem:     unit.DataSize(2 * 1024 * 1024 * 1024),
		MemSwap: unit.DataSize(2 * 1024 * 1024 * 1024),
	})
	assert.Empty(t, skipped)
	assert.Equal(t, update.Memory, update.MemorySwap)
}

// Docker rejects a swap limit below the memory limit. Sending it anyway would fail
// the whole update, taking the CPU limit down with it.
func Test_buildResourceUpdate_RejectsSwapBelowMemory(t *testing.T) {
	update, skipped := buildResourceUpdate(&entity.ImageBuildResourceSettings{
		CPUs:    2,
		Mem:     unit.DataSize(4 * 1024 * 1024 * 1024),
		MemSwap: unit.DataSize(1 * 1024 * 1024 * 1024),
	})

	assert.Len(t, skipped, 1)
	assert.Contains(t, skipped[0], "must be at least")
	assert.Zero(t, update.Memory, "the invalid pair must not be sent")
	assert.Zero(t, update.MemorySwap)
	assert.NotZero(t, update.NanoCPUs, "cpu is independent and must still be applied")
}

// A swap limit without a memory limit is meaningless to docker, which requires both.
func Test_buildResourceUpdate_SwapWithoutMemory(t *testing.T) {
	update, skipped := buildResourceUpdate(&entity.ImageBuildResourceSettings{
		MemSwap: unit.DataSize(1 * 1024 * 1024 * 1024),
	})

	assert.Len(t, skipped, 1)
	assert.Contains(t, skipped[0], "requires a mem limit")
	assert.Zero(t, update.MemorySwap)
}

func Test_buildResourceUpdate_CPUOnly(t *testing.T) {
	update, skipped := buildResourceUpdate(&entity.ImageBuildResourceSettings{CPUs: 3})

	assert.Empty(t, skipped)
	assert.EqualValues(t, 3*docker.UnitCPUNano, update.NanoCPUs)
	assert.Zero(t, update.Memory, "memory must be left alone when it is not configured")
	assert.Zero(t, update.MemorySwap)
}
