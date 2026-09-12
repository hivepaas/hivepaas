package appservice

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func testApp() *entity.App {
	return &entity.App{ID: "app-1", ProjectID: "proj-1", ProjectEnvID: "env-1"}
}

func TestWithAppLogLabelsStampsIdentity(t *testing.T) {
	got := WithAppLogLabels(map[string]string{"team": "a"}, testApp())

	assert.Equal(t, map[string]string{
		"team":               "a",
		LabelLogAppID:        "app-1",
		LabelLogProjectID:    "proj-1",
		LabelLogProjectEnvID: "env-1",
	}, got)
}

// A clone's ContainerSpec is a shallow copy of its source's. Stamping in place
// would write the clone's identity onto the source app as well.
func TestWithAppLogLabelsDoesNotTouchItsInput(t *testing.T) {
	source := map[string]string{LabelLogAppID: "source-app"}

	_ = WithAppLogLabels(source, testApp())

	assert.Equal(t, "source-app", source[LabelLogAppID])
}

// Copied from a source app, the old identity must not survive into the clone.
func TestWithAppLogLabelsReplacesAnInheritedIdentity(t *testing.T) {
	got := WithAppLogLabels(map[string]string{LabelLogAppID: "source-app"}, testApp())

	assert.Equal(t, "app-1", got[LabelLogAppID])
}

func TestIsLogDriverCollectible(t *testing.T) {
	assert.True(t, IsLogDriverCollectible(nil), "no driver is the daemon default, json-file")
	assert.True(t, IsLogDriverCollectible(&swarm.Driver{}))
	assert.True(t, IsLogDriverCollectible(&swarm.Driver{Name: "json-file"}))
	assert.False(t, IsLogDriverCollectible(&swarm.Driver{Name: "local"}),
		"local writes a binary format and no *-json.log")
	assert.False(t, IsLogDriverCollectible(&swarm.Driver{Name: "syslog"}))
	assert.False(t, IsLogDriverCollectible(&swarm.Driver{Name: "none"}))
}

func TestWithLogLabelsOptionNamesIdentityLabels(t *testing.T) {
	got := WithLogLabelsOption(&swarm.Driver{
		Name:    "json-file",
		Options: map[string]string{"max-size": "50m"},
	})

	assert.Equal(t, "hivepaas.app.id,hivepaas.project.id,hivepaas.projectEnv.id", got.Options["labels"])
	assert.Equal(t, "50m", got.Options["max-size"], "the operator's own options survive")
}

func TestWithLogLabelsOptionKeepsOperatorLabels(t *testing.T) {
	got := WithLogLabelsOption(&swarm.Driver{
		Name:    "json-file",
		Options: map[string]string{"labels": "team, hivepaas.app.id"},
	})

	assert.Equal(t, "hivepaas.app.id,hivepaas.project.id,hivepaas.projectEnv.id,team", got.Options["labels"])
}

// An operator who chose syslog keeps it untouched; the option would do nothing.
func TestWithLogLabelsOptionLeavesOtherDriversAlone(t *testing.T) {
	in := &swarm.Driver{Name: "syslog", Options: map[string]string{"syslog-address": "udp://x"}}

	assert.Same(t, in, WithLogLabelsOption(in))
	assert.Nil(t, WithLogLabelsOption(nil))
}

func TestWithLogLabelsOptionDoesNotTouchItsInput(t *testing.T) {
	in := &swarm.Driver{Name: "json-file", Options: map[string]string{"max-size": "1m"}}

	_ = WithLogLabelsOption(in)

	assert.NotContains(t, in.Options, "labels")
}

func TestHasLogIdentity(t *testing.T) {
	labeled := WithLogLabelsOption(&swarm.Driver{Name: "json-file"})
	cs := &swarm.ContainerSpec{Labels: map[string]string{LabelLogAppID: "app-1"}}

	assert.True(t, HasLogIdentity(labeled, cs, "app-1"))
	assert.False(t, HasLogIdentity(labeled, cs, "app-2"), "a clone still carrying its source's id")
	assert.False(t, HasLogIdentity(&swarm.Driver{Name: "json-file"}, cs, "app-1"), "no labels option")
	assert.False(t, HasLogIdentity(labeled, &swarm.ContainerSpec{}, "app-1"), "created before the labels")
	assert.False(t, HasLogIdentity(nil, cs, "app-1"), "daemon default carries no labels option")
	assert.False(t, HasLogIdentity(&swarm.Driver{Name: "local"}, cs, "app-1"))
}

// The default has to be something the collector reads and that names the
// identity labels, with limits, or apps would go uncollected or grow unbounded.
func TestDefaultLogDriverIsCollectibleAndBounded(t *testing.T) {
	d := DefaultLogDriver()

	assert.True(t, IsLogDriverCollectible(d))
	assert.True(t, HasLogIdentity(d, &swarm.ContainerSpec{Labels: map[string]string{LabelLogAppID: "a"}}, "a"))
	assert.Equal(t, "50m", d.Options["max-size"])
	assert.Equal(t, "20", d.Options["max-file"])
}

func TestDefaultLogDriverIsANewValueEachTime(t *testing.T) {
	DefaultLogDriver().Options["max-size"] = "changed"

	assert.Equal(t, "50m", DefaultLogDriver().Options["max-size"])
}
