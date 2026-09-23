package specserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) buildService(_ context.Context, state *buildState) error {
	applyService(state.req.Doc.Deployment.Service, state.req.Spec)
	return nil
}

// applyService writes the service block the way the service settings screen
// does. The mode changes only when the block gives one; the placement is the
// block's constraints after the ones HivePaaS derived, which stay.
func applyService(svc *specmodel.Service, spec *swarm.ServiceSpec) {
	if svc == nil {
		svc = &specmodel.Service{}
	}
	applyServiceMode(svc.ModeSpec, spec)

	task := &spec.TaskTemplate
	constraints := derivedConstraints(task.Placement, spec.Labels[labelAppPlacementConstraints])
	var preferences []swarm.PlacementPreference
	if p := svc.Placement; p != nil {
		constraints = append(constraints, p.Constraints...)
		for _, pref := range p.Preferences {
			if pref != nil && pref.Name == "spread" {
				preferences = append(preferences, swarm.PlacementPreference{
					Spread: &swarm.SpreadOver{SpreadDescriptor: pref.Value},
				})
			}
		}
	}
	if task.Placement == nil {
		task.Placement = &swarm.Placement{}
	}
	task.Placement.Constraints, task.Placement.Preferences = constraints, preferences
}

func applyServiceMode(m *specmodel.ServiceModeSpec, spec *swarm.ServiceSpec) {
	if m == nil || m.Mode == "" {
		return
	}
	spec.Mode = swarm.ServiceMode{}
	switch m.Mode {
	case docker.ServiceModeReplicated:
		spec.Mode.Replicated = &swarm.ReplicatedService{Replicas: copyUint(m.ServiceReplicas)}
	case docker.ServiceModeReplicatedJob:
		spec.Mode.ReplicatedJob = &swarm.ReplicatedJob{
			MaxConcurrent: copyUint(m.JobMaxConcurrent), TotalCompletions: copyUint(m.JobTotalCompletions),
		}
	case docker.ServiceModeGlobal:
		spec.Mode.Global = &swarm.GlobalService{}
	case docker.ServiceModeGlobalJob:
		spec.Mode.GlobalJob = &swarm.GlobalJob{}
	}
}

func copyUint(v *uint64) *uint64 {
	if v == nil {
		return nil
	}
	return new(*v)
}
