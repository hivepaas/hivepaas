package appservice

import (
	"slices"

	"github.com/moby/moby/api/types/swarm"
)

// ConstraintAppStopped is how a global service is stopped.
//
// Swarm refuses to change a service's mode, so a global service cannot be scaled
// to zero replicas the way a replicated one is. It is constrained to a node that
// cannot exist instead - a node id never contains a hyphen - and every task goes.
// Applying placement settings carries it over: it is not one of the constraints
// placement manages, and placement keeps those it does not manage.
const ConstraintAppStopped = "node.id==hivepaas-app-stopped"

// StopGlobalService pins spec to no node. It reports whether spec changed.
func StopGlobalService(spec *swarm.ServiceSpec) bool {
	if spec.TaskTemplate.Placement == nil {
		spec.TaskTemplate.Placement = &swarm.Placement{}
	}
	placement := spec.TaskTemplate.Placement
	if slices.Contains(placement.Constraints, ConstraintAppStopped) {
		return false
	}
	placement.Constraints = append(slices.Clone(placement.Constraints), ConstraintAppStopped)
	return true
}

// StartGlobalService undoes StopGlobalService. It reports whether spec changed.
func StartGlobalService(spec *swarm.ServiceSpec) bool {
	placement := spec.TaskTemplate.Placement
	if placement == nil {
		return false
	}
	i := slices.Index(placement.Constraints, ConstraintAppStopped)
	if i < 0 {
		return false
	}
	placement.Constraints = slices.Delete(slices.Clone(placement.Constraints), i, i+1)
	return true
}

// ForgetStoppedState drops what stopping an app wrote onto spec, for a spec that
// is about to become a different service: the state belongs to the service being
// replaced, so starting the app later must not restore it.
func ForgetStoppedState(spec *swarm.ServiceSpec) {
	delete(spec.Labels, LabelAppPrevServiceMode)
	StartGlobalService(spec)
}
