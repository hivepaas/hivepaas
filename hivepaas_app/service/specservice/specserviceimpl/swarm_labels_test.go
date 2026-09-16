package specserviceimpl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterUserLabelsRemovesEverythingRegenerated(t *testing.T) {
	kept := filterUserLabels(map[string]string{
		// HivePaaS rewrites these whenever the service is applied
		"hivepaas.app.info":                 `{"name":"a1"}`,
		"hivepaas.app.id":                   "app_1",
		"hivepaas.app.placementConstraints": "node.role==manager",
		// Docker's own, written by docker stack deploy
		"com.docker.stack.namespace": "p1",
		// Docker Desktop injects these, and they carry absolute host paths
		"desktop.docker.io/mounts/0/Source": "/Users/tnt/go/src/github.com/hivepaas/hivepaas/.appdata",
		"desktop.docker.io/mounts/0/Target": "/var/lib/hivepaas",
		// Traefik regenerates its own labels from the routing settings
		"traefik.http.routers.app.rule": "Host(`x.com`)",
		// except the ones a user wrote by hand
		"traefik.http.routers.x-custom-router-acme.entrypoints": "web",
		// and anything that is plainly the user's
		"team":        "platform",
		"cost-center": "eng",
	})

	assert.Equal(t, map[string]string{
		"traefik.http.routers.x-custom-router-acme.entrypoints": "web",
		"team":        "platform",
		"cost-center": "eng",
	}, kept)
}

func TestFilterUserLabelsHandlesNil(t *testing.T) {
	assert.Empty(t, filterUserLabels(nil))
}

func TestFilterUserConstraintsRemovesOnlyTheOnesHivePaaSAdded(t *testing.T) {
	kept := filterUserConstraints(
		[]string{
			"node.role == manager",
			"node.labels.hivepaas.role == control-plane",
			"node.labels.zone == eu",
		},
		"node.role==manager,node.labels.hivepaas.role==control-plane",
	)
	assert.Equal(t, []string{"node.labels.zone == eu"}, kept)
}

// With no managed-constraints label, every constraint is the user's.
func TestFilterUserConstraintsKeepsAllWhenNothingWasManaged(t *testing.T) {
	kept := filterUserConstraints([]string{"node.labels.zone == eu"}, "")
	assert.Equal(t, []string{"node.labels.zone == eu"}, kept)
}

func TestFilterUserConstraintsHandlesEmptyInput(t *testing.T) {
	assert.Nil(t, filterUserConstraints(nil, "node.role==manager"))
	assert.Nil(t, filterUserConstraints([]string{"node.role == manager"}, "node.role==manager"))
}

// Spacing must not decide the outcome: placement_apply.go compares constraints
// with their spacing removed, and so must this.
func TestFilterUserConstraintsIgnoresSpacing(t *testing.T) {
	kept := filterUserConstraints(
		[]string{"node.role   ==   manager", "node.labels.zone==eu"},
		"node.role == manager",
	)
	assert.Equal(t, []string{"node.labels.zone==eu"}, kept)
}

// The constant is duplicated from placementserviceimpl, which cannot export it
// without exposing an internal. Duplication is fine; drifting is not, and
// nothing else would notice: a mismatch means the exporter silently stops
// recognizing HivePaaS's own constraints and starts exporting them as the
// user's.
func TestPlacementConstraintsLabelMatchesPlacementService(t *testing.T) {
	assert.Equal(t, "hivepaas.app.placementConstraints", labelAppPlacementConstraints)

	source, err := os.ReadFile(
		"../../placementservice/placementserviceimpl/placement_apply.go")
	assert.NoError(t, err)
	assert.Contains(t, string(source),
		`labelAppPlacementConstraints = "`+labelAppPlacementConstraints+`"`,
		"placementserviceimpl declares a different value; the two have drifted")
}
