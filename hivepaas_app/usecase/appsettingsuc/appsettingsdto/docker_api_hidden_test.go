package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

func TestTheStorageScreenDoesNotShowTheSocket(t *testing.T) {
	data := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"}
	mounts, err := TransformStorageMounts(&StorageSettingsTransformInput{
		App: &entity.App{Key: "web", Project: &entity.Project{Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
		Service: &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{data, dockerapiservice.SocketMount("app1")}},
		}}},
		MountKeyCalculator: func(m *mount.Mount) string { return m.Target },
	})
	assert.NoError(t, err)
	if assert.Len(t, mounts, 1) {
		assert.Equal(t, "/data", mounts[0].Target)
	}
}

func TestTheNetworkScreenDoesNotShowTheAppsOwnNetwork(t *testing.T) {
	attachments, err := TransformNetworkAttachments(
		[]swarm.NetworkAttachmentConfig{{Target: "n-project"}, {Target: "n-dapi"}},
		&NetworkTransformationInput{DockerNetworks: map[string]*network.Summary{
			"n-project": {Network: network.Network{ID: "n-project", Name: "shop_prod_net"}},
			"n-dapi":    {Network: network.Network{ID: "n-dapi", Name: dockerapiservice.NetworkName("app1")}},
		}})
	assert.NoError(t, err)
	if assert.Len(t, attachments, 1) {
		assert.Equal(t, "shop_prod_net", attachments[0].Name)
	}
}
