package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// clientVisibleDetail renders the error the way BaseHandler.RenderError does
// for a caller outside the dev environment: Build's Detail field, not
// Error() (which is only the internal ERR_xxx identity) and not WithMsgLog's
// DebugLog (which RenderError strips before the response ever leaves dev).
func clientVisibleDetail(t *testing.T, err error) string {
	t.Helper()

	hpErr, ok := err.(hperrors.HPError)
	require.True(t, ok, "expected an hperrors.HPError, got %T", err)
	return hpErr.Build("en").Detail
}

// A bind volume is mounted by path, so the mount carries everything the node
// needs without consulting any daemon.
func TestBindMountTargetFromSetting(t *testing.T) {
	vol := &entity.ClusterVolume{
		Managed: true,
		Driver:  "local",
		DriverOpts: map[string]string{
			"type": "none", "device": "/srv/data", "o": "bind,rw,rshared",
		},
	}

	directory, propagation, ok := bindMountTarget(vol, "shop/prod/web")

	assert.True(t, ok)
	assert.Equal(t, "/srv/data/shop/prod/web", directory)
	assert.Equal(t, mount.PropagationRShared, propagation)
}

func TestBindMountTargetRejectsNonBindVolumes(t *testing.T) {
	tests := []struct {
		name string
		vol  *entity.ClusterVolume
	}{
		{"nfs", &entity.ClusterVolume{Driver: "local", DriverOpts: map[string]string{"type": "nfs", "device": ":/e"}}},
		{"no device", &entity.ClusterVolume{Driver: "local", DriverOpts: map[string]string{"type": "none"}}},
		{"custom driver", &entity.ClusterVolume{Driver: "some-plugin"}},
		{"discovered, nothing recorded", &entity.ClusterVolume{}},
		{"nil", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := bindMountTarget(tt.vol, "shop/prod/web")
			assert.False(t, ok)
		})
	}
}

// Driver config in the mount spec is what makes a node materialize the volume
// correctly instead of inventing an empty one.
func TestApplyVolumeDriverConfig(t *testing.T) {
	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "01JVOL"}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
		Labels:     map[string]string{"hivepaas.volume.name": "shared"},
	}

	applyVolumeDriverConfig(dockerMnt, vol)

	assert.NotNil(t, dockerMnt.VolumeOptions)
	assert.Equal(t, "local", dockerMnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", dockerMnt.VolumeOptions.DriverConfig.Options["device"])
	assert.Equal(t, "shared", dockerMnt.VolumeOptions.Labels["hivepaas.volume.name"])
}

// A volume HivePaaS did not author is mounted by name and nothing is inferred
// about it: stamping a specification we did not write would be a guess.
func TestApplyVolumeDriverConfigSkipsUnmanaged(t *testing.T) {
	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "someones-volume"}
	vol := &entity.ClusterVolume{
		Managed:    false,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs"},
	}

	applyVolumeDriverConfig(dockerMnt, vol)

	assert.Nil(t, dockerMnt.VolumeOptions)
}

// A client that names its own driver is overriding, not omitting: the setting
// must not clobber a choice the caller made on purpose.
func TestApplyVolumeDriverConfigUnlessOverriddenKeepsClientDriverConfig(t *testing.T) {
	dockerMnt := &mount.Mount{
		Type: mount.TypeVolume,
		VolumeOptions: &mount.VolumeOptions{
			DriverConfig: &mount.Driver{Name: "nfs", Options: map[string]string{"device": ":/client/chosen"}},
		},
	}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data"},
	}

	applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)

	assert.Equal(t, "nfs", dockerMnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/client/chosen", dockerMnt.VolumeOptions.DriverConfig.Options["device"])
}

// The client can set Subpath/NoCopy/Labels without naming a driver; the setting
// fills in the driver config it omitted without touching what it did set.
func TestApplyVolumeDriverConfigUnlessOverriddenFillsInWhenClientOmitsDriverConfig(t *testing.T) {
	dockerMnt := &mount.Mount{
		Type: mount.TypeVolume,
		VolumeOptions: &mount.VolumeOptions{
			Subpath: "shop/prod/web",
			NoCopy:  true,
			Labels:  map[string]string{"app": "web"},
		},
	}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data"},
	}

	applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)

	assert.Equal(t, "local", dockerMnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", dockerMnt.VolumeOptions.DriverConfig.Options["device"])
	assert.Equal(t, "shop/prod/web", dockerMnt.VolumeOptions.Subpath)
	assert.True(t, dockerMnt.VolumeOptions.NoCopy)
	assert.Equal(t, "web", dockerMnt.VolumeOptions.Labels["app"])
}

// The common case: the client sends no VolumeOptions at all, so the mount has
// none to preserve and the setting's specification is what makes it work.
func TestApplyVolumeDriverConfigUnlessOverriddenAppliesWhenNoVolumeOptions(t *testing.T) {
	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "01JVOL"}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data"},
	}

	applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)

	assert.NotNil(t, dockerMnt.VolumeOptions)
	assert.Equal(t, "local", dockerMnt.VolumeOptions.DriverConfig.Name)
}

// Unmanaged still wins even when the client's VolumeOptions leave room for a
// DriverConfig to be filled in: a volume we did not author is not ours to
// describe, guard or no guard.
func TestApplyVolumeDriverConfigUnlessOverriddenSkipsUnmanaged(t *testing.T) {
	dockerMnt := &mount.Mount{
		Type: mount.TypeVolume,
		VolumeOptions: &mount.VolumeOptions{
			Subpath: "shop/prod/web",
		},
	}
	vol := &entity.ClusterVolume{
		Managed:    false,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data"},
	}

	applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)

	assert.Nil(t, dockerMnt.VolumeOptions.DriverConfig)
}

// newVolumeSetting builds a TypeVolume/TypeCluster-matchable *entity.Setting
// the way a row read back from the database looks: SetData caches the parsed
// struct on the receiver, but AsClusterVolume (via VolumePinsForMounts) must
// still work from a Setting that only carries the raw Data string, since that
// is what a fresh read from storage produces.
func newVolumeSetting(t *testing.T, id, refID, name, nodeID string) *entity.Setting {
	t.Helper()

	setting := &entity.Setting{ID: id, RefID: refID, Type: base.SettingTypeClusterVolume, Name: name}
	require.NoError(t, setting.SetData(&entity.ClusterVolume{NodeID: nodeID}))
	return &entity.Setting{
		ID: setting.ID, Name: setting.Name, RefID: setting.RefID, Type: setting.Type, Data: setting.Data,
	}
}

// newBindVolumeSetting builds a bind-matchable *entity.Setting the same way -
// matched by its recorded device rather than by RefID, since that is all a
// rewritten bind mount carries by the time it comes back from docker.
func newBindVolumeSetting(t *testing.T, id, name, nodeID, device string) *entity.Setting {
	t.Helper()

	setting := &entity.Setting{ID: id, Type: base.SettingTypeClusterVolume, Name: name}
	require.NoError(t, setting.SetData(&entity.ClusterVolume{
		NodeID:     nodeID,
		DriverOpts: map[string]string{"type": "none", "device": device},
	}))
	return &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
}

// This is the case the original plan's Task 7 missed: it only looked at the
// volumes behind mounts being added, and an unchanged existing mount (a
// TypeVolume mount here, matched by id) never shows up there at all.
func TestRefuseConflictingVolumePinsCatchesAnUnchangedMountAgainstANewOne(t *testing.T) {
	pgdata := newVolumeSetting(t, "vol-pgdata", "ref-pgdata", "pgdata", "node-1")
	uploads := newBindVolumeSetting(t, "vol-uploads", "uploads", "node-2", "/srv/uploads")

	mounts := []mount.Mount{
		{Type: mount.TypeVolume, Source: "ref-pgdata"},               // unchanged, carried over from the load phase
		{Type: mount.TypeBind, Source: "/srv/uploads/shop/prod/web"}, // new bind mount
	}

	err := refuseConflictingVolumePins(mounts, []*entity.Setting{pgdata, uploads})

	detail := clientVisibleDetail(t, err)
	assert.Contains(t, detail, "pgdata")
	assert.Contains(t, detail, "uploads")
}

func TestRefuseConflictingVolumePinsCatchesTwoBindMountsOnDifferentNodes(t *testing.T) {
	a := newBindVolumeSetting(t, "vol-a", "a", "node-1", "/srv/a")
	b := newBindVolumeSetting(t, "vol-b", "b", "node-2", "/srv/b")

	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/a/web"},
		{Type: mount.TypeBind, Source: "/srv/b/web"},
	}

	err := refuseConflictingVolumePins(mounts, []*entity.Setting{a, b})

	detail := clientVisibleDetail(t, err)
	assert.Contains(t, detail, "'a'")
	assert.Contains(t, detail, "'b'")
}

func TestRefuseConflictingVolumePinsAllowsPinsOnTheSameNode(t *testing.T) {
	a := newBindVolumeSetting(t, "vol-a", "a", "node-1", "/srv/a")
	b := newBindVolumeSetting(t, "vol-b", "b", "node-1", "/srv/b")

	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/a/web"},
		{Type: mount.TypeBind, Source: "/srv/b/web"},
	}

	assert.NoError(t, refuseConflictingVolumePins(mounts, []*entity.Setting{a, b}))
}

func TestRefuseConflictingVolumePinsAllowsAPinnedVolumeAlongsideAnUnpinnedOne(t *testing.T) {
	pinned := newBindVolumeSetting(t, "vol-a", "a", "node-1", "/srv/a")
	unpinned := newBindVolumeSetting(t, "vol-b", "b", "", "/srv/b")

	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/a/web"},
		{Type: mount.TypeBind, Source: "/srv/b/web"},
	}

	assert.NoError(t, refuseConflictingVolumePins(mounts, []*entity.Setting{pinned, unpinned}))
}

func TestRefuseConflictingVolumePinsAllowsNoPinnedVolumes(t *testing.T) {
	a := newBindVolumeSetting(t, "vol-a", "a", "", "/srv/a")
	b := newBindVolumeSetting(t, "vol-b", "b", "", "/srv/b")

	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/a/web"},
		{Type: mount.TypeBind, Source: "/srv/b/web"},
	}

	assert.NoError(t, refuseConflictingVolumePins(mounts, []*entity.Setting{a, b}))
}

func TestRefuseConflictingVolumePinsAllowsABindMountThatMatchesNoVolume(t *testing.T) {
	pinned := newBindVolumeSetting(t, "vol-a", "a", "node-1", "/srv/a")

	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/a/web"},
		{Type: mount.TypeBind, Source: "/unrelated/path"},
	}

	assert.NoError(t, refuseConflictingVolumePins(mounts, []*entity.Setting{pinned}))
}
