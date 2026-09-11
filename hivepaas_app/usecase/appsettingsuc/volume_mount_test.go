package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

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
