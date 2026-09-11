package placementservice

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// newBindVolume builds the *entity.Setting the way a row read back from the
// database looks: SetData caches the parsed struct on the receiver, but
// AsClusterVolume must still work from a Setting that only carries the raw
// Data string, since that is what a fresh read from storage produces.
func newBindVolume(t *testing.T, name, nodeID, device string) *entity.Setting {
	t.Helper()

	setting := &entity.Setting{ID: name, Type: base.SettingTypeClusterVolume, Name: name}
	assert.NoError(t, setting.SetData(&entity.ClusterVolume{
		NodeID:     nodeID,
		Managed:    true,
		DriverOpts: map[string]string{"type": "none", "device": device},
	}))
	return &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
}

func newVolumeMount(refID, name, nodeID string) *entity.Setting {
	setting := &entity.Setting{ID: name, RefID: refID, Type: base.SettingTypeClusterVolume, Name: name}
	setting.MustSetData(&entity.ClusterVolume{NodeID: nodeID})
	return &entity.Setting{
		ID: setting.ID, Name: setting.Name, RefID: setting.RefID, Type: setting.Type, Data: setting.Data,
	}
}

func bindMount(source string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: source}
}

// This is the case Task 6b exists for: useBindMountIfAppropriate rewrites a
// pinned local/type=none volume into a plain bind mount before it ever reaches
// docker, so the mount carries nothing but a host path. Finding the pin behind
// it is the whole point of this file.
func TestVolumePinsForMountsFindsAPinnedVolumeBehindABindMount(t *testing.T) {
	web := newBindVolume(t, "web", "node-1", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/shop/prod/web")},
		[]*entity.Setting{web},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "web", NodeID: "node-1"}}, pins)
}

func TestVolumePinsForMountsMatchesTypeVolumeByRefID(t *testing.T) {
	pgdata := newVolumeMount("vol-ref-1", "pgdata", "node-1")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{{Type: mount.TypeVolume, Source: "vol-ref-1"}},
		[]*entity.Setting{pgdata},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}}, pins)
}

func TestVolumePinsForMountsMatchesTypeClusterByRefID(t *testing.T) {
	pgdata := newVolumeMount("vol-ref-1", "pgdata", "node-1")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{{Type: mount.TypeCluster, Source: "vol-ref-1"}},
		[]*entity.Setting{pgdata},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}}, pins)
}

// A device is a directory, not a string prefix: /srv/data must not swallow a
// sibling directory that merely starts with the same characters.
func TestVolumePinsForMountsRespectsThePathBoundary(t *testing.T) {
	data := newBindVolume(t, "data", "node-1", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/database/x")},
		[]*entity.Setting{data},
	)

	assert.NoError(t, err)
	assert.Empty(t, pins)
}

// Two volumes can legitimately be nested (a whole-tree volume and a more
// specific one carved out underneath it); the more specific device is the
// better answer for a source under both.
func TestVolumePinsForMountsPrefersTheLongestDevice(t *testing.T) {
	outer := newBindVolume(t, "outer", "node-1", "/srv/data")
	inner := newBindVolume(t, "inner", "node-2", "/srv/data/pg")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/pg/x")},
		[]*entity.Setting{outer, inner},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "inner", NodeID: "node-2"}}, pins)
}

// A trailing slash on the recorded device is cosmetic and must not defeat the
// match.
func TestVolumePinsForMountsToleratesATrailingSlashOnTheDevice(t *testing.T) {
	web := newBindVolume(t, "web", "node-1", "/srv/data/")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/x")},
		[]*entity.Setting{web},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "web", NodeID: "node-1"}}, pins)
}

// A bind unrelated to any known volume, and a mount type that carries no
// volume identity at all, both contribute nothing.
func TestVolumePinsForMountsIgnoresUnmatchedBindsAndOtherMountTypes(t *testing.T) {
	web := newBindVolume(t, "web", "node-1", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{
			bindMount("/unrelated/path"),
			{Type: mount.TypeTmpfs},
		},
		[]*entity.Setting{web},
	)

	assert.NoError(t, err)
	assert.Empty(t, pins)
}

// A volume HivePaaS did not author never became a bind mount in the first place
// - bindMountTarget refuses it - so a bind source that happens to sit under its
// recorded device came from somewhere else entirely. Matching it would pin the
// service to a node on the strength of a coincidence of paths.
func TestVolumePinsForMountsIgnoresAnUnmanagedVolumeBehindABindMount(t *testing.T) {
	discovered := &entity.Setting{ID: "d", Type: base.SettingTypeClusterVolume, Name: "discovered"}
	assert.NoError(t, discovered.SetData(&entity.ClusterVolume{
		NodeID:     "node-1",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data"},
	}))
	discovered = &entity.Setting{
		ID: discovered.ID, Name: discovered.Name, Type: discovered.Type, Data: discovered.Data,
	}

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/shop/prod/web")},
		[]*entity.Setting{discovered},
	)

	assert.NoError(t, err)
	assert.Empty(t, pins)
}

// The managed volume is the one the bind came from, even when an unmanaged
// setting records a longer device under the same source - specificity only
// ranks candidates, it does not turn a volume we did not author into one.
func TestVolumePinsForMountsPrefersAManagedVolumeOverALongerUnmanagedDevice(t *testing.T) {
	managed := newBindVolume(t, "managed", "node-1", "/srv/data")

	unmanaged := &entity.Setting{ID: "u", Type: base.SettingTypeClusterVolume, Name: "unmanaged"}
	assert.NoError(t, unmanaged.SetData(&entity.ClusterVolume{
		NodeID:     "node-2",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data/shop"},
	}))
	unmanaged = &entity.Setting{
		ID: unmanaged.ID, Name: unmanaged.Name, Type: unmanaged.Type, Data: unmanaged.Data,
	}

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/shop/prod/web")},
		[]*entity.Setting{managed, unmanaged},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "managed", NodeID: "node-1"}}, pins)
}

// A setting written before Task 1 started recording DriverOpts has nothing to
// compare a bind's source against. That is a gap for a later backfill task to
// close, not an error this one should raise.
func TestVolumePinsForMountsDoesNotErrorOnALegacySettingWithNoDriverOpts(t *testing.T) {
	legacy := &entity.Setting{ID: "legacy", Type: base.SettingTypeClusterVolume, Name: "legacy"}
	legacy.MustSetData(&entity.ClusterVolume{NodeID: "node-1"})
	legacy = &entity.Setting{ID: legacy.ID, Name: legacy.Name, Type: legacy.Type, Data: legacy.Data}

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/x")},
		[]*entity.Setting{legacy},
	)

	assert.NoError(t, err)
	assert.Empty(t, pins)
}

// Two mounts pointing at the same volume (by ref id, or by two bind paths
// under the same device) must still produce one pin - the same volume cannot
// impose the same constraint twice, and a naive implementation could return it
// twice and let VolumePinConstraint see a false "conflict" between a pin and
// itself.
func TestVolumePinsForMountsDedupesAVolumeMountedTwice(t *testing.T) {
	web := newBindVolume(t, "web", "node-1", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{
			bindMount("/srv/data/a"),
			bindMount("/srv/data/b"),
		},
		[]*entity.Setting{web},
	)

	assert.NoError(t, err)
	assert.Equal(t, []VolumePin{{VolumeName: "web", NodeID: "node-1"}}, pins)
}

// Two settings recording the identical longest device are, literally, two
// descriptions of the same directory. VolumePinsForMounts does not pick a
// winner between them - it returns both pins and leaves the disagreement for
// VolumePinConstraint to report as a conflict, which is the only honest thing
// to do when two volumes claim the same data lives on two different nodes.
func TestVolumePinsForMountsReturnsBothPinsOnATiedDeviceThatConflicts(t *testing.T) {
	a := newBindVolume(t, "a", "node-1", "/srv/data")
	b := newBindVolume(t, "b", "node-2", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/x")},
		[]*entity.Setting{a, b},
	)

	assert.NoError(t, err)
	assert.ElementsMatch(t, []VolumePin{
		{VolumeName: "a", NodeID: "node-1"},
		{VolumeName: "b", NodeID: "node-2"},
	}, pins)

	_, conflict := VolumePinConstraint(pins)
	assert.NotNil(t, conflict,
		"two volumes recording the same device but pinned to different nodes is a genuine conflict")
}

// The same tie, but this time the two settings agree on the node - both pins
// still come back (VolumePinsForMounts still does not pick a winner), and
// together they resolve to the single constraint they agree on.
func TestVolumePinsForMountsReturnsBothPinsOnATiedDeviceThatAgree(t *testing.T) {
	a := newBindVolume(t, "a", "node-1", "/srv/data")
	b := newBindVolume(t, "b", "node-1", "/srv/data")

	pins, err := VolumePinsForMounts(
		[]mount.Mount{bindMount("/srv/data/x")},
		[]*entity.Setting{a, b},
	)

	assert.NoError(t, err)
	assert.ElementsMatch(t, []VolumePin{
		{VolumeName: "a", NodeID: "node-1"},
		{VolumeName: "b", NodeID: "node-1"},
	}, pins)

	constraint, conflict := VolumePinConstraint(pins)
	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}
