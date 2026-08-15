package placementserviceimpl

import (
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"
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
		prevHivepaasConstraints = strings.Split(raw, ",")
	}

	for _, constraint := range currConstraints {
		if !gofn.Contain(prevHivepaasConstraints, constraint) {
			finalConstraints = append(finalConstraints, constraint)
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
			if label == "" {
				continue
			}

			parts := strings.SplitN(label, "=", 2) //nolint:mnd
			key := strings.TrimSpace(parts[0])
			if key == "" {
				continue
			}
			val := ""
			if len(parts) == 2 { //nolint:mnd
				val = strings.TrimSpace(parts[1])
			}

			if val != "" {
				newHivepaasConstraints = append(newHivepaasConstraints, fmt.Sprintf("node.labels.%s!=%s", key, val))
			} else {
				newHivepaasConstraints = append(newHivepaasConstraints, fmt.Sprintf("node.labels.%s!=true", key))
			}
		}
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
