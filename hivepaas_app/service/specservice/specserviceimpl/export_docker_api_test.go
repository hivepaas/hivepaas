package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// The socket and the network are this installation's, named after this app's
// id. Import gives an app with the setting its own.
func TestExportLeavesOutTheAppsSocketAndNetwork(t *testing.T) {
	task := &swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{
			{Type: mount.TypeTmpfs, Target: "/scratch"},
			dockerapiservice.SocketMount("app1"),
		}},
		Networks: []swarm.NetworkAttachmentConfig{{Target: "n-project"}, {Target: "n-dapi"}},
	}

	storage, err := mapAppStorage(task, nil, func(string) (string, *specmodel.ExternalRef) { return "", nil })
	assert.NoError(t, err)
	if assert.NotNil(t, storage) {
		assert.Len(t, storage.DockerMounts, 1)
		assert.Contains(t, storage.DockerMounts, "/scratch")
	}

	networks := mapNetworks(task, nil, map[string]string{
		"n-project": "shop_prod_net", "n-dapi": dockerapiservice.NetworkName("app1"),
	})
	if assert.NotNil(t, networks) && assert.Len(t, networks.Attachments, 1) {
		assert.Equal(t, "shop_prod_net", networks.Attachments[0].Name)
	}
}
