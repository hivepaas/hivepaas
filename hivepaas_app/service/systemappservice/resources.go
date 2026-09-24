package systemappservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/docker"
)

// nanoPerCPU is how swarm counts cores.
const nanoPerCPU = 1e9

// Resources are the limits a system app runs under. Zero in a field is no cap.
//
// They are not stored with the feature's settings: the app's service is where
// they live - the app's own resource screen reads and writes the same place - so
// a settings page reads them off the service and writes them back there.
type Resources struct {
	// CPULimit is in cores.
	CPULimit float64
	// MemoryLimit is in bytes.
	MemoryLimit int64
}

// ResourcesOf is what spec runs under.
func ResourcesOf(spec *swarm.ServiceSpec) Resources {
	var res Resources
	if spec.TaskTemplate.Resources == nil || spec.TaskTemplate.Resources.Limits == nil {
		return res
	}
	limits := spec.TaskTemplate.Resources.Limits
	res.CPULimit = float64(limits.NanoCPUs) / nanoPerCPU
	res.MemoryLimit = limits.MemoryBytes
	return res
}

// ApplyResources writes res onto spec the way an app's document does, and says
// whether that changed anything. The OOM priority goes only with a memory limit:
// see base.OomScoreAdjSystemAddon.
func ApplyResources(spec *swarm.ServiceSpec, res Resources) bool {
	var nanoCPUs int64
	if res.CPULimit > 0 {
		nanoCPUs = docker.TruncateCPUsAsNano(res.CPULimit, docker.MinCPUFraction)
	}
	var oomScoreAdj int64
	if res.MemoryLimit > 0 {
		oomScoreAdj = base.OomScoreAdjSystemAddon
	}

	container := spec.TaskTemplate.ContainerSpec
	var currentNanoCPUs, currentMemory int64
	if limits := spec.TaskTemplate.Resources; limits != nil && limits.Limits != nil {
		currentNanoCPUs, currentMemory = limits.Limits.NanoCPUs, limits.Limits.MemoryBytes
	}
	if currentNanoCPUs == nanoCPUs && currentMemory == res.MemoryLimit && container.OomScoreAdj == oomScoreAdj {
		return false
	}

	if spec.TaskTemplate.Resources == nil {
		spec.TaskTemplate.Resources = &swarm.ResourceRequirements{}
	}
	if spec.TaskTemplate.Resources.Limits == nil {
		spec.TaskTemplate.Resources.Limits = &swarm.Limit{}
	}
	spec.TaskTemplate.Resources.Limits.NanoCPUs = nanoCPUs
	spec.TaskTemplate.Resources.Limits.MemoryBytes = res.MemoryLimit
	container.OomScoreAdj = oomScoreAdj
	return true
}
