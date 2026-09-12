package placementserviceimpl

import (
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

const (
	labelAppPlacementConstraints = "hivepaas.app.placementConstraints"
)

//nolint:gocognit
func (s *service) applyPlacementSettings(
	data *placementSettingsData,
) {
	spec := &data.Service.Spec
	var finalConstraints []string
	var currConstraints []string

	// Read current constraints
	if spec.TaskTemplate.Placement != nil {
		currConstraints = spec.TaskTemplate.Placement.Constraints
	}

	// Keep user-set constraints (filter out constraints previously managed by HivePaaS)
	var prevHivepaasConstraints []string
	if raw, ok := spec.Labels[labelAppPlacementConstraints]; ok && raw != "" {
		for _, item := range strings.Split(raw, ",") {
			k, op, v := dockerhelper.ParsePlacementConstraint(item)
			if op != "" {
				prevHivepaasConstraints = append(prevHivepaasConstraints, k+op+v)
			} else if trimmed := strings.TrimSpace(item); trimmed != "" {
				prevHivepaasConstraints = append(prevHivepaasConstraints, trimmed)
			}
		}
	}

	for _, constraint := range currConstraints {
		k, op, v := dockerhelper.ParsePlacementConstraint(constraint)
		constraintNorm := constraint
		if op != "" {
			constraintNorm = k + op + v
		}
		if !gofn.Contain(prevHivepaasConstraints, constraintNorm) {
			finalConstraints = append(finalConstraints, constraintNorm)
		}
	}

	// Build new HivePaaS constraints
	var newHivepaasConstraints []string

	if data.PlacementSettings.ExcludeManagerNodes {
		newHivepaasConstraints = append(newHivepaasConstraints, "node.role!=manager")
	}

	if data.PlacementSettings.ExcludeBuildNodes && data.BuildSettings != nil {
		for _, nodeID := range data.BuildSettings.Workers.NodeIDs {
			if nodeID == "" {
				continue
			}
			newHivepaasConstraints = append(newHivepaasConstraints, "node.id!="+nodeID)
		}

		for _, label := range data.BuildSettings.Workers.NodeLabels {
			if constraint, ok := dockerhelper.NodeLabelConstraint(label, "!="); ok {
				newHivepaasConstraints = append(newHivepaasConstraints, constraint)
			}
		}
	}

	// The operator's own label rules. A selector that cannot become a
	// constraint was refused when the settings were saved; one that arrives
	// here anyway is skipped rather than emitted half-formed.
	for _, label := range data.PlacementSettings.RequireNodeLabels {
		if constraint, ok := dockerhelper.NodeLabelConstraint(label, "=="); ok {
			newHivepaasConstraints = append(newHivepaasConstraints, constraint)
		}
	}
	for _, label := range data.PlacementSettings.ExcludeNodeLabels {
		if constraint, ok := dockerhelper.NodeLabelConstraint(label, "!="); ok {
			newHivepaasConstraints = append(newHivepaasConstraints, constraint)
		}
	}

	// The one required constraint HivePaaS emits.
	constraint, conflict := placementservice.VolumePinConstraint(data.VolumePins)
	switch {
	case conflict != nil:
		// UpdateAppStorageSettings refuses a contradictory pin set before it is
		// saved, but that is the only door with a lock on it: a spec written
		// before that check existed, or edited on the service directly, still
		// arrives here. Emitting no constraint stays the only honest answer - no
		// node satisfies both pins - but a service that has quietly lost its pin
		// must not also be invisible.
		s.logger.Warnf("app %s mounts volumes with contradictory node pins, so no placement "+
			"constraint is applied and its tasks may be scheduled anywhere: %v",
			data.Service.Spec.Name, conflict)
	case constraint != "":
		newHivepaasConstraints = append(newHivepaasConstraints, constraint)
	}

	finalConstraints = append(finalConstraints, newHivepaasConstraints...)
	if spec.TaskTemplate.Placement == nil {
		spec.TaskTemplate.Placement = &swarm.Placement{}
	}
	spec.TaskTemplate.Placement.Constraints = finalConstraints
	data.HasChanges = !gofn.ContentEqual(finalConstraints, currConstraints)

	// Store new hivepaas-set constraints in labels
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	if len(newHivepaasConstraints) > 0 {
		spec.Labels[labelAppPlacementConstraints] = strings.Join(newHivepaasConstraints, ",")
	} else {
		delete(spec.Labels, labelAppPlacementConstraints)
	}
}
