package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// A mount of the app's own directory in a volume names the volume, which a data
// backup picks; another app's directory, or a mount no volume accounts for,
// names none.
func TestAStorageMountNamesTheVolumeOfTheAppsOwnDirectory(t *testing.T) {
	mounts := []mount.Mount{
		{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"},
		{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/other"},
		{Type: mount.TypeBind, Source: "/srv/x", Target: "/x"},
	}
	resp, err := TransformStorageMounts(&StorageSettingsTransformInput{
		App: &entity.App{Key: "web", Project: &entity.Project{Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
		Service: &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Mounts: mounts},
		}}},
		MountKeyCalculator: func(m *mount.Mount) string { return m.Target },
		MountDescs: []*volumeservice.AppMountDesc{
			{AppKey: "web", Own: true, VolumeID: "vol1"},
			{AppKey: "db", VolumeID: "vol1"},
			{},
		},
	})

	assert.NoError(t, err)
	if assert.Len(t, resp, 3) {
		assert.Equal(t, "vol1", resp[0].VolumeID)
		assert.Empty(t, resp[1].VolumeID)
		assert.Empty(t, resp[2].VolumeID)
	}
}

// A mount in a volume HivePaaS made is shown by the volume's name, whoever's
// directory it is: its source is the id docker knows it by.
func TestAStorageMountCarriesItsVolumesName(t *testing.T) {
	mounts := []mount.Mount{
		{Type: mount.TypeVolume, Source: "01M4AB3ZQGPW5T21RVB5JVFG6M", Target: "/data"},
		{Type: mount.TypeBind, Source: "/srv/x", Target: "/x"},
	}
	resp, err := TransformStorageMounts(&StorageSettingsTransformInput{
		App: &entity.App{Key: "web", Project: &entity.Project{Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
		Service: &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Mounts: mounts},
		}}},
		MountKeyCalculator: func(m *mount.Mount) string { return m.Target },
		MountDescs:         []*volumeservice.AppMountDesc{{VolumeName: "uploads"}, {}},
	})

	assert.NoError(t, err)
	if assert.Len(t, resp, 2) {
		assert.Equal(t, "uploads", resp[0].SourceName)
		assert.Equal(t, "01M4AB3ZQGPW5T21RVB5JVFG6M", resp[0].Source, "the source stays what it is")
		assert.Empty(t, resp[1].SourceName)
	}
}
