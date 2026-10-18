package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
)

func serviceWithArtifact(t *testing.T) (*swarm.Service, mount.Mount) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{HostDir: "/var/lib/hivepaas"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
	data := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"}
	walg := artifactservice.Mount(artifactservice.CacheHostDir()+"/aa", "/opt/hivepaas/bin/wal-g")
	return &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{data, walg}},
	}}}, data
}

// Artifacts are the engine's, not chosen on the storage screen: it does not
// show them, and they do not move its version.
func TestTheStorageScreenDoesNotShowArtifacts(t *testing.T) {
	service, data := serviceWithArtifact(t)

	mounts, err := TransformStorageMounts(&StorageSettingsTransformInput{
		App: &entity.App{Key: "db", Project: &entity.Project{Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
		Service:            service,
		MountKeyCalculator: func(m *mount.Mount) string { return m.Target },
	})

	assert.NoError(t, err)
	if assert.Len(t, mounts, 1) {
		assert.Equal(t, "/data", mounts[0].Target)
	}
	withoutArtifact := &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{data}},
	}}}
	assert.Equal(t, StorageSettingsVersion(withoutArtifact), StorageSettingsVersion(service))
}
