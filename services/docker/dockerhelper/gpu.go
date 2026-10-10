package dockerhelper

import (
	"slices"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/services/docker"
)

func isGPU(res swarm.GenericResource) bool {
	return res.DiscreteResourceSpec != nil && docker.IsGPUKind(res.DiscreteResourceSpec.Kind) ||
		res.NamedResourceSpec != nil && docker.IsGPUKind(res.NamedResourceSpec.Kind)
}

func oneGPU() swarm.GenericResource {
	return swarm.GenericResource{DiscreteResourceSpec: &swarm.DiscreteGenericResource{
		Kind: docker.GenericResourceGPU, Value: 1,
	}}
}

func reservedResources(task *swarm.TaskSpec) []swarm.GenericResource {
	if task == nil || task.Resources == nil || task.Resources.Reservations == nil {
		return nil
	}
	return task.Resources.Reservations.GenericResources
}

// HasLegacyGPU says whether the task has the capability Enable GPU used to
// write, which docker refuses.
func HasLegacyGPU(task *swarm.TaskSpec) bool {
	return task != nil && task.ContainerSpec != nil &&
		slices.Contains(task.ContainerSpec.CapabilityAdd, docker.CapabilityGPU)
}

// ReservesGPU says whether the task has a GPU: one reserved among its generic
// resources, of any maker, or the capability Enable GPU used to write.
func ReservesGPU(task *swarm.TaskSpec) bool {
	return slices.ContainsFunc(reservedResources(task), isGPU) || HasLegacyGPU(task)
}

// ShownGenericResources are the generic resources the task reserves, as a
// screen or a document shows them: with the capability Enable GPU used to
// write read as the one NVIDIA GPU it asked for, which saving them reserves.
func ShownGenericResources(task *swarm.TaskSpec) []swarm.GenericResource {
	resources := slices.Clone(reservedResources(task))
	if HasLegacyGPU(task) && !slices.ContainsFunc(resources, isGPU) {
		resources = append(resources, oneGPU())
	}
	return resources
}

// ReserveOneGPU reserves one NVIDIA GPU for the task - what enableGPU asks -
// unless it reserves a GPU already.
func ReserveOneGPU(task *swarm.TaskSpec) {
	if task == nil || slices.ContainsFunc(reservedResources(task), isGPU) {
		return
	}
	if task.Resources == nil {
		task.Resources = &swarm.ResourceRequirements{}
	}
	if task.Resources.Reservations == nil {
		task.Resources.Reservations = &swarm.Resources{}
	}
	task.Resources.Reservations.GenericResources = append(task.Resources.Reservations.GenericResources, oneGPU())
}

// DropLegacyGPU takes away the capability Enable GPU used to write: docker
// refuses the container that has it.
func DropLegacyGPU(task *swarm.TaskSpec) {
	if task == nil || task.ContainerSpec == nil {
		return
	}
	task.ContainerSpec.CapabilityAdd = slices.DeleteFunc(slices.Clone(task.ContainerSpec.CapabilityAdd),
		func(capability string) bool { return capability == docker.CapabilityGPU })
}
