// Package specserviceimpl builds configuration spec bundles.
//
// It reads the Docker Swarm service directly rather than through
// appsettingsdto.Transform*. Those DTOs are the dashboard's wire shape and are
// free to change when the UI does; a spec is a durable artifact read back by
// versions that do not exist yet. ARCHITECTURE.md also places dto above
// service, so the dependency would run the wrong way.
package specserviceimpl

import (
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// labelAppPlacementConstraints is where HivePaaS records the placement
// constraints it added itself, so the user's can be told apart from them. It
// must match placementserviceimpl's constant of the same value.
const labelAppPlacementConstraints = "hivepaas.app.placementConstraints"

// managedLabelPrefixes are rewritten whenever the service is applied, so a spec
// carrying them would disagree with the system the moment it was imported.
//
// This is a maintained denylist rather than a single prefix check because third
// parties write labels onto services too: desktop.docker.io is Docker Desktop's,
// not HivePaaS's, and it embeds absolute paths from the machine that ran the
// export. The next such writer will not be called hivepaas either.
var managedLabelPrefixes = []string{
	"hivepaas.",
	"com.docker.stack.",
	"desktop.docker.io/",
}

// traefikCustomMarker marks the Traefik labels a user wrote by hand. Traefik
// regenerates every other traefik.* label from the routing settings - the same
// rule updateSwarmServiceLabels uses when it cleans up its own.
const traefikCustomMarker = ".x-custom-"

// filterUserLabels keeps only the labels a person put there.
func filterUserLabels(labels map[string]string) map[string]string {
	kept := make(map[string]string, len(labels))
	for key, value := range labels {
		if isManagedLabel(key) {
			continue
		}
		kept[key] = value
	}
	// None kept is no labels at all, which is what a document read back from
	// YAML holds too: an empty map is not written.
	if len(kept) == 0 {
		return nil
	}
	return kept
}

func isManagedLabel(key string) bool {
	for _, prefix := range managedLabelPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return strings.HasPrefix(key, "traefik.") && !strings.Contains(key, traefikCustomMarker)
}

// filterUserConstraints removes the placement constraints HivePaaS derived from
// app-placement settings and from volume node pinning, leaving the user's.
//
// managed is the raw value of the labelAppPlacementConstraints label. Without
// this filter an import duplicates every constraint, and carries a node.id
// belonging to a node that does not exist - leaving the app schedulable nowhere.
func filterUserConstraints(constraints []string, managed string) []string {
	if len(constraints) == 0 {
		return nil
	}

	managedSet := managedConstraintSet(managed)

	kept := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		if gofn.Contain(managedSet, normalizeConstraint(constraint)) {
			continue
		}
		kept = append(kept, constraint)
	}
	if len(kept) == 0 {
		return nil
	}
	return kept
}

// normalizeConstraint strips the spacing, the way placement_apply.go does when
// it compares constraints.
func normalizeConstraint(constraint string) string {
	k, op, v := dockerhelper.ParsePlacementConstraint(constraint)
	if op == "" {
		return strings.TrimSpace(constraint)
	}
	return k + op + v
}

// managedConstraintSet is the placement constraints HivePaaS derived, as the
// label listing them says, each normalized.
func managedConstraintSet(managed string) []string {
	set := make([]string, 0, 4) //nolint:mnd
	for _, item := range strings.Split(managed, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			set = append(set, normalizeConstraint(trimmed))
		}
	}
	return set
}

// derivedConstraints is the constraints of a placement HivePaaS derived: the ones
// export leaves out, and building the block keeps.
func derivedConstraints(placement *swarm.Placement, managed string) []string {
	if placement == nil {
		return nil
	}
	set := managedConstraintSet(managed)
	var out []string
	for _, constraint := range placement.Constraints {
		if gofn.Contain(set, normalizeConstraint(constraint)) {
			out = append(out, constraint)
		}
	}
	return out
}
