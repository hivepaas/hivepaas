package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/logging"
)

// serviceUpdateRetryMax is how often a resource change is retried when the
// service moved on under it.
const serviceUpdateRetryMax = 2

// nanoPerCPU is how swarm counts cores.
const nanoPerCPU = 1e9

// defaultBackendResources is what a backend gets when nothing says otherwise.
func defaultBackendResources() logging.Resources {
	return logging.Resources{MemoryLimit: logging.DefaultBackendMemoryLimit.Bytes()}
}

// resourcesOf is what a service runs under, the way the settings screen shows it.
func resourcesOf(spec *swarm.ServiceSpec) logging.Resources {
	var res logging.Resources
	if spec.TaskTemplate.Resources == nil || spec.TaskTemplate.Resources.Limits == nil {
		return res
	}
	limits := spec.TaskTemplate.Resources.Limits
	res.CPULimit = float64(limits.NanoCPUs) / nanoPerCPU
	res.MemoryLimit = limits.MemoryBytes
	return res
}

// applyBackendResources writes the limits onto the backend's service. The
// service is where they live - the app's own resource screen reads and writes
// the same place - so nothing else holds them.
func (s *service) applyBackendResources(ctx context.Context, app *entity.App, res logging.Resources) error {
	err := s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, svc *swarm.Service) (bool, error) {
			return setResources(&svc.Spec, res), nil
		}, serviceUpdateRetryMax, 0)
	return hperrors.Wrap(err)
}

// setResources writes res onto spec the way the app's document does, and says
// whether that changed anything. The OOM priority goes only with a memory
// limit: see base.OomScoreAdjSystemAddon.
func setResources(spec *swarm.ServiceSpec, res logging.Resources) bool {
	var nanoCPUs int64
	if res.CPULimit > 0 {
		nanoCPUs = docker.TruncateCPUsAsNano(res.CPULimit, docker.MinCPUFraction)
	}
	var oomScoreAdj int64
	if res.MemoryLimit > 0 {
		oomScoreAdj = base.OomScoreAdjSystemAddon
	}

	container := spec.TaskTemplate.ContainerSpec
	current := resourcesOf(spec)
	if spec.TaskTemplate.Resources != nil && spec.TaskTemplate.Resources.Limits != nil &&
		spec.TaskTemplate.Resources.Limits.NanoCPUs == nanoCPUs &&
		current.MemoryLimit == res.MemoryLimit && container.OomScoreAdj == oomScoreAdj {
		return false
	}
	if nanoCPUs == 0 && res.MemoryLimit == 0 && current == (logging.Resources{}) &&
		container.OomScoreAdj == oomScoreAdj {
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
