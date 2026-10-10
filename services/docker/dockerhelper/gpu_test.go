package dockerhelper

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/docker"
)

func gpus(count int64) swarm.GenericResource {
	return swarm.GenericResource{DiscreteResourceSpec: &swarm.DiscreteGenericResource{
		Kind: docker.GenericResourceGPU, Value: count,
	}}
}

// Enable GPU is one GPU reserved: swarm places the app on a node with one
// free. The capability it wrote before - which docker refuses - goes.
func TestEnableGPUReservesOneGPUAndDropsTheCapabilityItWroteBefore(t *testing.T) {
	task := &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		CapabilityAdd: []string{"NET_ADMIN", docker.CapabilityGPU},
	}}
	assert.True(t, ReservesOneGPU(task), "an app saved the old way has it")

	SetOneGPU(task, true)
	assert.Equal(t, []string{"NET_ADMIN"}, task.ContainerSpec.CapabilityAdd)
	assert.Equal(t, []swarm.GenericResource{gpus(1)}, task.Resources.Reservations.GenericResources)
	assert.True(t, ReservesOneGPU(task))
	assert.Empty(t, WithoutOneGPU(task.Resources.Reservations.GenericResources), "shown as Enable GPU")

	SetOneGPU(task, true)
	assert.Len(t, task.Resources.Reservations.GenericResources, 1, "once")
}

// GPUs reserved by hand are the app's: Enable GPU adds none beside them, and
// they are not Enable GPU's one.
func TestGPUsReservedByHandAreKept(t *testing.T) {
	task := &swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{},
		Resources: &swarm.ResourceRequirements{Reservations: &swarm.Resources{
			GenericResources: []swarm.GenericResource{gpus(2)},
		}},
	}
	SetOneGPU(task, true)
	assert.Equal(t, []swarm.GenericResource{gpus(2)}, task.Resources.Reservations.GenericResources)
	assert.False(t, ReservesOneGPU(task))
	assert.True(t, ReservesGPU(task))
	assert.Equal(t, []swarm.GenericResource{gpus(2)}, WithoutOneGPU(task.Resources.Reservations.GenericResources))

	SetOneGPU(task, false)
	assert.Equal(t, []swarm.GenericResource{gpus(2)}, task.Resources.Reservations.GenericResources)
	assert.False(t, ReservesGPU(&swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}))
}
