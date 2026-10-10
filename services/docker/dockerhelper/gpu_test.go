package dockerhelper

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/docker"
)

func reserved(kind string, count int64) swarm.GenericResource {
	return swarm.GenericResource{DiscreteResourceSpec: &swarm.DiscreteGenericResource{Kind: kind, Value: count}}
}

func reserving(resources ...swarm.GenericResource) *swarm.TaskSpec {
	return &swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{},
		Resources:     &swarm.ResourceRequirements{Reservations: &swarm.Resources{GenericResources: resources}},
	}
}

// A GPU is one reserved among the generic resources, of any maker.
func TestAGPUIsOneReservedOfAnyMaker(t *testing.T) {
	assert.True(t, ReservesGPU(reserving(reserved("NVIDIA-GPU", 1))))
	assert.True(t, ReservesGPU(reserving(reserved("AMD_GPU", 2))))
	assert.False(t, ReservesGPU(reserving(reserved("SSD", 1))))
	assert.False(t, ReservesGPU(&swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}))
}

// The capability Enable GPU used to write is shown as the NVIDIA GPU it asked
// for, and goes when the app is saved; enableGPU reserves one NVIDIA GPU,
// unless the app has a GPU already.
func TestTheCapabilityEnableGPUUsedToWriteIsShownAsTheGPUItAskedFor(t *testing.T) {
	task := &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		CapabilityAdd: []string{"NET_ADMIN", docker.CapabilityGPU},
	}}
	assert.True(t, ReservesGPU(task))
	assert.Equal(t, []swarm.GenericResource{reserved(docker.GenericResourceGPU, 1)}, ShownGenericResources(task))

	DropLegacyGPU(task)
	assert.Equal(t, []string{"NET_ADMIN"}, task.ContainerSpec.CapabilityAdd)
	assert.Empty(t, ShownGenericResources(task))

	ReserveOneGPU(task)
	ReserveOneGPU(task)
	assert.Equal(t, []swarm.GenericResource{reserved(docker.GenericResourceGPU, 1)},
		task.Resources.Reservations.GenericResources, "once")

	amd := reserving(reserved("AMD_GPU", 1))
	ReserveOneGPU(amd)
	assert.Equal(t, []swarm.GenericResource{reserved("AMD_GPU", 1)}, amd.Resources.Reservations.GenericResources,
		"a GPU of the app's own is kept, and none added")
}
