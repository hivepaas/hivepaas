package placementserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

// driftingNodes are the nodes running tasks that a pin to pinnedNodeID would
// move them off.
//
// Only an id pin can be checked this way. A label pin names a set of nodes, and
// a node id on its own cannot say whether it is in that set - so an empty
// pinnedNodeID reports nothing rather than a mismatch it cannot substantiate.
func driftingNodes(tasks []swarm.Task, pinnedNodeID string) []string {
	if pinnedNodeID == "" {
		return nil
	}

	seen := map[string]struct{}{}
	var nodes []string
	for _, task := range tasks {
		if task.Status.State != swarm.TaskStateRunning || task.NodeID == "" {
			continue
		}
		if task.NodeID == pinnedNodeID {
			continue
		}
		if _, found := seen[task.NodeID]; found {
			continue
		}
		seen[task.NodeID] = struct{}{}
		nodes = append(nodes, task.NodeID)
	}
	return nodes
}

// warnOnPinDrift says when applying the pin will move a task that is running.
//
// It only reports. Moving the task is what the constraint does, and it happens
// when the service spec is next written - which is a moment the operator chose,
// rather than a sweep that relocates running apps on upgrade.
func (s *service) warnOnPinDrift(ctx context.Context, data *placementSettingsData) {
	constraint, conflict := placementservice.VolumePinConstraint(data.VolumePins)
	if conflict != nil || constraint == "" {
		// A conflict is refused rather than applied, and no constraint moves
		// nothing - in neither case is a task about to go anywhere.
		return
	}

	// The first pin carrying a node id is the one the constraint was built from:
	// a constraint agreed on by every pinned volume (which is what a nil
	// conflict means) cannot name one node id here and a different one - or a
	// label - there.
	var pinnedNodeID string
	for _, pin := range data.VolumePins {
		if pin.NodeID != "" {
			pinnedNodeID = pin.NodeID
			break
		}
	}
	if pinnedNodeID == "" || data.Service.ID == "" {
		// A label pin has no single node to compare against, and a service with
		// no id has never been created, so it has no tasks to move.
		return
	}

	// The filter is on desired state; driftingNodes then checks the state the
	// task is actually in. A task swarm wants running but which is still
	// starting up is not yet sitting on data anywhere.
	resp, err := s.dockerManager.ServiceTaskList(ctx, data.Service.ID,
		[]swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		// The warning is a courtesy on the way to applying the settings. Failing
		// the apply because docker would not list tasks would trade a missing
		// note for a broken deploy.
		return
	}
	nodes := driftingNodes(resp.Items, pinnedNodeID)
	if len(nodes) == 0 {
		return
	}
	s.logger.Warnf(
		"app %s runs on %v but mounts a volume pinned to %s; applying the pin moves it, "+
			"and the data it has been using stays where it is",
		data.Service.Spec.Name, nodes, pinnedNodeID)
}
