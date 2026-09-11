package volumedto

import (
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// newStoredClusterVolumeSetting builds a setting the way a row read back from
// the database looks. MustSetData caches the parsed struct on the Setting and
// AsClusterVolume returns that cache before ever reading Data, so a setting a
// test stored into and read straight back would assert nothing about what
// TransformVolume does with data that actually went through JSON.
func newStoredClusterVolumeSetting(data *entity.ClusterVolume) *entity.Setting {
	setting := &entity.Setting{ID: "set_1", RefID: "vol_1", Type: base.SettingTypeClusterVolume}
	setting.MustSetData(data)
	return &entity.Setting{ID: setting.ID, RefID: setting.RefID, Type: setting.Type, Data: setting.Data}
}

func TestTransformVolumeUsesRecordedSpecWhenDockerHasNotBuiltItYet(t *testing.T) {
	t.Run("bind volume", func(t *testing.T) {
		setting := newStoredClusterVolumeSetting(&entity.ClusterVolume{
			Driver:     string(docker.VolumeDriverLocal),
			DriverOpts: map[string]string{"type": "none", "device": "/data/pg", "o": "bind,rw"},
		})

		resp, err := TransformVolume(setting, nil, entity.NewRefClusterObjects())

		assert.NoError(t, err)
		assert.Equal(t, docker.VolumeDriverLocal, resp.Driver)
		assert.NotNil(t, resp.BindOptions)
		assert.Equal(t, "/data/pg", resp.BindOptions.Directory)
		assert.Nil(t, resp.NfsOptions)
		assert.Nil(t, resp.TmpfsOptions)
		// Nothing has told HivePaaS where the volume actually lives yet.
		assert.Empty(t, resp.Mountpoint)
	})

	t.Run("nfs volume", func(t *testing.T) {
		setting := newStoredClusterVolumeSetting(&entity.ClusterVolume{
			Driver:     string(docker.VolumeDriverLocal),
			DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
		})

		resp, err := TransformVolume(setting, nil, entity.NewRefClusterObjects())

		assert.NoError(t, err)
		assert.NotNil(t, resp.NfsOptions)
		assert.Equal(t, ":/exports/data", resp.NfsOptions.Device)
		assert.Nil(t, resp.BindOptions)
		assert.Nil(t, resp.TmpfsOptions)
		assert.Empty(t, resp.Mountpoint)
	})

	t.Run("tmpfs volume", func(t *testing.T) {
		setting := newStoredClusterVolumeSetting(&entity.ClusterVolume{
			Driver:     string(docker.VolumeDriverLocal),
			DriverOpts: map[string]string{"type": "tmpfs", "o": "size=100m,uid=1000"},
		})

		resp, err := TransformVolume(setting, nil, entity.NewRefClusterObjects())

		assert.NoError(t, err)
		assert.NotNil(t, resp.TmpfsOptions)
		assert.Equal(t, 1000, resp.TmpfsOptions.UID)
		assert.Nil(t, resp.BindOptions)
		assert.Nil(t, resp.NfsOptions)
		assert.Empty(t, resp.Mountpoint)
	})
}

// When docker has actually built the volume, its values win over the setting's
// spec even where the two agree, because docker describes something real.
func TestTransformVolumeDockerWinsOverRecordedSpec(t *testing.T) {
	setting := newStoredClusterVolumeSetting(&entity.ClusterVolume{
		Driver:     string(docker.VolumeDriverLocal),
		DriverOpts: map[string]string{"type": "none", "device": "/data/pg-spec", "o": "bind,rw"},
		Labels:     map[string]string{"spec": "yes"},
	})

	refClusterObjects := entity.NewRefClusterObjects()
	refClusterObjects.AddRefVolumes(volume.Volume{
		Name:       "vol_1",
		Driver:     string(docker.VolumeDriverLocal),
		Mountpoint: "/var/lib/docker/volumes/vol_1/_data",
		CreatedAt:  "2026-01-02T15:04:05Z",
		Options:    map[string]string{"type": "none", "device": "/data/pg-real", "o": "bind,ro"},
		Labels:     map[string]string{"real": "yes"},
		UsageData:  &volume.UsageData{RefCount: 2, Size: 4096},
	})

	resp, err := TransformVolume(setting, nil, refClusterObjects)

	assert.NoError(t, err)
	assert.Equal(t, "/var/lib/docker/volumes/vol_1/_data", resp.Mountpoint)
	assert.False(t, resp.CreatedAt.IsZero())
	assert.Equal(t, int64(2), resp.RefCount)
	assert.Equal(t, int64(4096), resp.Size)
	assert.Equal(t, map[string]string{"real": "yes"}, resp.Labels)
	assert.NotNil(t, resp.BindOptions)
	assert.Equal(t, "/data/pg-real", resp.BindOptions.Directory)
	assert.True(t, resp.BindOptions.Readonly)
}

// A volume with neither a recorded spec nor a live docker volume - an old
// setting from before Task 1 - must not panic and must not invent values.
func TestTransformVolumeWithNeitherSourceInventsNothing(t *testing.T) {
	setting := newStoredClusterVolumeSetting(&entity.ClusterVolume{})

	resp, err := TransformVolume(setting, nil, entity.NewRefClusterObjects())

	assert.NoError(t, err)
	assert.Empty(t, resp.Driver)
	assert.Nil(t, resp.BindOptions)
	assert.Nil(t, resp.NfsOptions)
	assert.Nil(t, resp.TmpfsOptions)
	assert.Empty(t, resp.Mountpoint)
}
