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
