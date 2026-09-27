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
