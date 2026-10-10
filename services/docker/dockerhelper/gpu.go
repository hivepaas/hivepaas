package dockerhelper

import (
	"slices"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/services/docker"
)

// isOneGPU is the reservation Enable GPU stands for: one GPU, by count.
func isOneGPU(res swarm.GenericResource) bool {
	return res.DiscreteResourceSpec != nil && res.DiscreteResourceSpec.Kind == docker.GenericResourceGPU &&
		res.DiscreteResourceSpec.Value == 1
}

// isGPU is any reservation of GPUs: a count of them, or one by its name.
func isGPU(res swarm.GenericResource) bool {
	return res.DiscreteResourceSpec != nil && res.DiscreteResourceSpec.Kind == docker.GenericResourceGPU ||
		res.NamedResourceSpec != nil && res.NamedResourceSpec.Kind == docker.GenericResourceGPU
}

func reservedResources(task *swarm.TaskSpec) []swarm.GenericResource {
	if task == nil || task.Resources == nil || task.Resources.Reservations == nil {
		return nil
	}
	return task.Resources.Reservations.GenericResources
}

func hasLegacyGPU(task *swarm.TaskSpec) bool {
	return task != nil && task.ContainerSpec != nil &&
		slices.Contains(task.ContainerSpec.CapabilityAdd, docker.CapabilityGPU)
}

// ReservesOneGPU says whether the task has what Enable GPU gives it: one GPU
// reserved, or the capability Enable GPU wrote before.
func ReservesOneGPU(task *swarm.TaskSpec) bool {
	return slices.ContainsFunc(reservedResources(task), isOneGPU) || hasLegacyGPU(task)
}

// ReservesGPU says whether the task has a GPU in any way: Enable GPU's, or
// GPUs reserved by hand among its generic resources.
func ReservesGPU(task *swarm.TaskSpec) bool {
	return slices.ContainsFunc(reservedResources(task), isGPU) || hasLegacyGPU(task)
}

// WithoutOneGPU are the generic resources but the one GPU Enable GPU stands
// for, which a screen shows as Enable GPU instead.
func WithoutOneGPU(resources []swarm.GenericResource) []swarm.GenericResource {
	return slices.DeleteFunc(slices.Clone(resources), isOneGPU)
}

// SetOneGPU gives the task Enable GPU's one GPU, as a reservation, unless it
// reserves GPUs already; off, it leaves what the reservations say. Either way
// the capability Enable GPU wrote before goes: docker refuses the container
// that has it.
func SetOneGPU(task *swarm.TaskSpec, enable bool) {
	if task == nil {
		return
	}
	if task.ContainerSpec != nil {
		task.ContainerSpec.CapabilityAdd = slices.DeleteFunc(slices.Clone(task.ContainerSpec.CapabilityAdd),
			func(capability string) bool { return capability == docker.CapabilityGPU })
	}
	if !enable || slices.ContainsFunc(reservedResources(task), isGPU) {
		return
	}
	if task.Resources == nil {
		task.Resources = &swarm.ResourceRequirements{}
	}
	if task.Resources.Reservations == nil {
		task.Resources.Reservations = &swarm.Resources{}
	}
	task.Resources.Reservations.GenericResources = append(task.Resources.Reservations.GenericResources,
		swarm.GenericResource{DiscreteResourceSpec: &swarm.DiscreteGenericResource{
			Kind: docker.GenericResourceGPU, Value: 1,
		}})
}
