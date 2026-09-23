package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func blankTask() *swarm.TaskSpec {
	return &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}
}

// Every field export reads back out of a task's resources, built and read again.
func TestResourcesRoundTrip(t *testing.T) {
	swap, shm, swappiness := unit.DataSize(512<<20), unit.DataSize(64<<20), int64(10)
	want := &specmodel.Resources{
		Reservations: &specmodel.ResourceReservations{
			CPUs: 0.5, Memory: 256 << 20,
			GenericResources: []*specmodel.GenericResource{{Kind: "gpu", Value: "2"}, {Kind: "ssd", Value: "fast"}},
		},
		Limits: &specmodel.ResourceLimits{CPUs: 1, Memory: 512 << 20, Pids: 100},
		Memory: &specmodel.Memory{Swap: &swap, Swappiness: &swappiness, ShmSize: &shm},
		Capabilities: &specmodel.Capabilities{
			CapabilityAdd: []string{"NET_ADMIN"}, Sysctls: map[string]string{"vm.max_map_count": "262144"},
			Ulimits: []*specmodel.Ulimit{{Name: "nofile", Soft: 1024, Hard: 2048}}, OomScoreAdj: -100,
		},
	}
	task := blankTask()

	applyResources(want, task)

	assert.Equal(t, want, mapResources(task))
}

// A block the document leaves out clears what the task held, so building an
// exported document over a running service converges on the document.
func TestResourcesLeftOutAreCleared(t *testing.T) {
	task := blankTask()
	applyResources(&specmodel.Resources{
		Limits:       &specmodel.ResourceLimits{CPUs: 1},
		Memory:       &specmodel.Memory{ShmSize: new(unit.DataSize(64 << 20))},
		Capabilities: &specmodel.Capabilities{CapabilityAdd: []string{"NET_ADMIN"}},
	}, task)

	applyResources(nil, task)

	assert.Nil(t, mapResources(task))
	assert.Empty(t, task.ContainerSpec.Mounts, "the shared-memory mount goes with memory.shmSize")
}
