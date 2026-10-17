package clustercleanupserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type servicesDocker struct {
	docker.Manager
	services []swarm.Service
}

func (d *servicesDocker) ServiceList(context.Context, ...docker.ServiceListOption) (*client.ServiceListResult, error) {
	return &client.ServiceListResult{Items: d.services}, nil
}

func withMounts(mounts ...mount.Mount) *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Mounts: mounts}}}
}

// What a node keeps is what the swarm's services bind, as they run and as they
// ran before their last update - which a rollback goes back to.
func TestKeptArtifactsAreTheOnesServicesBind(t *testing.T) {
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{HostDir: "/var/lib/hivepaas"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
	cache := artifactservice.CacheHostDir()
	data := mount.Mount{Type: mount.TypeVolume, Source: "pgdata", Target: "/var/lib/postgresql"}
	s := &service{dockerManager: &servicesDocker{services: []swarm.Service{
		{Spec: *withMounts(data, artifactservice.Mount(cache+"/new", "/opt/hivepaas/bin/wal-g")),
			PreviousSpec: withMounts(data, artifactservice.Mount(cache+"/old", "/opt/hivepaas/bin/wal-g"))},
		{Spec: *withMounts(artifactservice.Mount(cache+"/new", "/opt/hivepaas/bin/wal-g"))},
		{Spec: swarm.ServiceSpec{}},
	}}}

	kept, err := s.KeptArtifacts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{"new", "old"}, kept)
}

func artifactsData(known bool, pruneImages bool) *clusterCleanupData {
	return &clusterCleanupData{
		ClusterCleanupReq: &clustercleanupservice.ClusterCleanupReq{
			CleanupSettings: &entity.SystemClusterCleanup{Enabled: true, PruneImages: pruneImages,
				GeneralRetention: timeutil.Duration(24 * time.Hour)},
			CleanupImages:  base.CleanupFlagTrue,
			KeepArtifacts:  []string{"kept"},
			ArtifactsKnown: known,
		},
		Output: &entity.ClusterNodeCleanupOutput{},
	}
}

func oldFiles(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	old := time.Now().Add(-72 * time.Hour)
	for _, name := range names {
		path := filepath.Join(dir, name)
		assert.NoError(t, os.WriteFile(path, []byte("x"), 0o755))
		assert.NoError(t, os.Chtimes(path, old, old))
	}
	return dir
}

// Artifacts go with images: the files a node keeps for what it runs.
func TestCleanupPrunesArtifactsWithImages(t *testing.T) {
	dir := oldFiles(t, "kept", "gone")
	data := artifactsData(true, true)

	assert.NoError(t, cleanupArtifactsIn(data, dir, time.Now()))

	assert.Equal(t, 1, data.Output.ArtifactsDeleted)
	assert.FileExists(t, filepath.Join(dir, "kept"))
	assert.NoFileExists(t, filepath.Join(dir, "gone"))
}

// A manager of an older version sends no list, and a run that does not prune
// images leaves artifacts too: nothing is removed without knowing what is used.
func TestCleanupLeavesArtifactsItCannotTellAbout(t *testing.T) {
	for name, data := range map[string]*clusterCleanupData{
		"an older manager":  artifactsData(false, true),
		"images not pruned": artifactsData(true, false),
	} {
		dir := oldFiles(t, "kept", "gone")

		assert.NoError(t, cleanupArtifactsIn(data, dir, time.Now()), name)

		assert.Zero(t, data.Output.ArtifactsDeleted, name)
		assert.FileExists(t, filepath.Join(dir, "gone"), name)
	}
}
