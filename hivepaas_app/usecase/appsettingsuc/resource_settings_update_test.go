package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func memoryOf(swap, shm unit.DataSize, swappiness int64) *appsettingsdto.Memory {
	return &appsettingsdto.Memory{Swap: &swap, ShmSize: &shm, Swappiness: &swappiness}
}

// Swap, swappiness and shared memory are what the settings say, and emptied,
// they are docker's defaults again: no swap of the app's own, no swappiness,
// no /dev/shm mount. The app's other mounts stay.
func TestTheMemorySettingsAreWhatTheySayAndEmptiedAreGone(t *testing.T) {
	data := &updateAppResourceSettingsData{Service: &swarm.Service{Spec: swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
			Mounts: []mount.Mount{{Type: mount.TypeVolume, Source: "data", Target: "/data"}},
		}},
	}}}
	uc := &UC{}

	uc.prepareUpdatingAppMemory(&appsettingsdto.UpdateAppResourceSettingsReq{
		Memory: memoryOf(32*unit.MB, 128*unit.MB, 10),
	}, data)
	task := &data.Service.Spec.TaskTemplate
	assert.Equal(t, int64(32*unit.MB), *task.Resources.SwapBytes)
	assert.Equal(t, int64(10), *task.Resources.MemorySwappiness)
	assert.Len(t, task.ContainerSpec.Mounts, 2)
	assert.Equal(t, int64(128*unit.MB), task.ContainerSpec.Mounts[1].TmpfsOptions.SizeBytes)

	for name, memory := range map[string]*appsettingsdto.Memory{
		"each emptied": {},
		"none given":   nil,
	} {
		uc.prepareUpdatingAppMemory(&appsettingsdto.UpdateAppResourceSettingsReq{
			Memory: memoryOf(32*unit.MB, 128*unit.MB, 10),
		}, data)
		uc.prepareUpdatingAppMemory(&appsettingsdto.UpdateAppResourceSettingsReq{Memory: memory}, data)
		assert.Nil(t, task.Resources.SwapBytes, name)
		assert.Nil(t, task.Resources.MemorySwappiness, name)
		assert.Equal(t, []mount.Mount{{Type: mount.TypeVolume, Source: "data", Target: "/data"}},
			task.ContainerSpec.Mounts, name)
	}

	// No swap at all is a value of its own, not an empty one.
	uc.prepareUpdatingAppMemory(&appsettingsdto.UpdateAppResourceSettingsReq{Memory: memoryOf(0, 0, 0)}, data)
	assert.Equal(t, int64(0), *task.Resources.SwapBytes)
}

// Enable GPU is shown from the GPU reserved, and saved as one: an app saved
// the old way - the capability docker refuses - is given the reservation in
// its place. A request saying nothing of the capabilities keeps it.
func TestEnableGPUIsAReservationOfOneGPU(t *testing.T) {
	legacy := &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{CapabilityAdd: []string{"SYS_PTRACE", docker.CapabilityGPU}},
	}}}
	shown, err := appsettingsdto.TransformResourceSettings(legacy)
	assert.NoError(t, err)
	assert.True(t, shown.Capabilities.EnableGPU)
	assert.Equal(t, []string{"SYS_PTRACE"}, shown.Capabilities.CapabilityAdd)

	data := &updateAppResourceSettingsData{Service: legacy}
	(&UC{}).prepareUpdatingAppResourceSettings(&appsettingsdto.UpdateAppResourceSettingsReq{
		Reservations: shown.Reservations, Limits: shown.Limits, Memory: shown.Memory, Capabilities: shown.Capabilities,
	}, data)
	task := &data.Service.Spec.TaskTemplate
	assert.Equal(t, []string{"SYS_PTRACE"}, task.ContainerSpec.CapabilityAdd)
	assert.Equal(t, []swarm.GenericResource{{DiscreteResourceSpec: &swarm.DiscreteGenericResource{
		Kind: docker.GenericResourceGPU, Value: 1,
	}}}, task.Resources.Reservations.GenericResources)

	shown, _ = appsettingsdto.TransformResourceSettings(data.Service)
	assert.True(t, shown.Capabilities.EnableGPU)
	assert.Empty(t, shown.Reservations.GenericResources, "the GPU is shown as Enable GPU")

	(&UC{}).prepareUpdatingAppResourceSettings(&appsettingsdto.UpdateAppResourceSettingsReq{}, data)
	assert.True(t, dockerhelper.ReservesOneGPU(task), "kept by a request saying nothing of it")

	(&UC{}).prepareUpdatingAppResourceSettings(&appsettingsdto.UpdateAppResourceSettingsReq{
		Capabilities: &appsettingsdto.Capabilities{},
	}, data)
	assert.False(t, dockerhelper.ReservesGPU(task), "turned off")
}
