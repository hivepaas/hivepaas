package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
)

// An artifact is a file in this installation's cache, bound by the engine on
// every deploy: a spec does not carry it, as it does not carry the socket.
func TestExportLeavesArtifactsOut(t *testing.T) {
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{HostDir: "/var/lib/hivepaas"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
	conf := mount.Mount{Type: mount.TypeBind, Source: "/srv/conf", Target: "/etc/app/config"}
	walg := artifactservice.Mount(artifactservice.CacheHostDir()+"/aa", "/opt/hivepaas/bin/wal-g")

	storage, err := mapAppStorage(&swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Mounts: []mount.Mount{conf, walg},
	}}, nil, nil)

	assert.NoError(t, err)
	assert.Contains(t, storage.DockerMounts, "/etc/app/config")
	assert.NotContains(t, storage.DockerMounts, "/opt/hivepaas/bin/wal-g")
}
