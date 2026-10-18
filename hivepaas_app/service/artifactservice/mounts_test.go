package artifactservice

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

func useHostDir(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{HostDir: "/var/lib/hivepaas"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func artifactMount(sha, target string) mount.Mount {
	return Mount("/var/lib/hivepaas/cache/artifacts/sha256/"+sha, target)
}

func TestAnArtifactMountIsAReadOnlyBindThatCreatesNothing(t *testing.T) {
	useHostDir(t)
	m := artifactMount("aa", "/opt/hivepaas/bin/wal-g")

	assert.Equal(t, mount.TypeBind, m.Type)
	assert.True(t, m.ReadOnly)
	assert.True(t, m.BindOptions == nil || !m.BindOptions.CreateMountpoint,
		"Docker must never make a directory where the file belongs")
	assert.True(t, IsArtifactMount(&m))
}

func TestIsArtifactMount(t *testing.T) {
	useHostDir(t)
	for name, m := range map[string]mount.Mount{
		"the app's own bind":     {Type: mount.TypeBind, Source: "/srv/data", Target: "/data"},
		"a volume":               {Type: mount.TypeVolume, Source: "pgdata", Target: "/var/lib/postgresql"},
		"another cache's bind":   {Type: mount.TypeBind, Source: "/var/lib/hivepaas/cache/app-templates/x", Target: "/x"},
		"the cache dir itself":   {Type: mount.TypeBind, Source: "/var/lib/hivepaas/cache/artifacts/sha256", Target: "/x"},
		"a volume of that name!": {Type: mount.TypeVolume, Source: "/var/lib/hivepaas/cache/artifacts/sha256/aa"},
	} {
		assert.False(t, IsArtifactMount(&m), name)
	}
}

// Attach replaces whatever artifacts the spec had - an older version's - and
// leaves the app's own mounts as they are.
func TestAttachReplacesTheArtifactMounts(t *testing.T) {
	useHostDir(t)
	data := mount.Mount{Type: mount.TypeVolume, Source: "pgdata", Target: "/var/lib/postgresql/data"}
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Mounts: []mount.Mount{data, artifactMount("old", "/opt/hivepaas/bin/wal-g")},
	}}}

	Attach(spec, []mount.Mount{artifactMount("new", "/opt/hivepaas/bin/wal-g")})

	assert.Equal(t, []mount.Mount{data, artifactMount("new", "/opt/hivepaas/bin/wal-g")},
		spec.TaskTemplate.ContainerSpec.Mounts)

	Attach(spec, nil)
	assert.Equal(t, []mount.Mount{data}, spec.TaskTemplate.ContainerSpec.Mounts, "none is none")

	empty := &swarm.ServiceSpec{}
	Attach(empty, []mount.Mount{artifactMount("new", "/opt/hivepaas/bin/wal-g")})
	assert.Len(t, empty.TaskTemplate.ContainerSpec.Mounts, 1)
}

// A mount at an artifact's target goes whatever its source: one of the cache's
// bound from a path this backend does not recognize - an agent whose
// HP_STORAGE_HOST_DIR was not set - would otherwise be kept beside the new one,
// and docker refuses a spec that mounts one path twice.
func TestAttachReplacesAnyMountAtAnArtifactsTarget(t *testing.T) {
	useHostDir(t)
	data := mount.Mount{Type: mount.TypeVolume, Source: "pgdata", Target: "/var/lib/postgresql/data"}
	stale := mount.Mount{Type: mount.TypeBind, Source: "cache/artifacts/sha256/old", Target: "/opt/hivepaas/bin/wal-g"}
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Mounts: []mount.Mount{data, stale},
	}}}

	Attach(spec, []mount.Mount{artifactMount("new", "/opt/hivepaas/bin/wal-g")})

	assert.Equal(t, []mount.Mount{data, artifactMount("new", "/opt/hivepaas/bin/wal-g")},
		spec.TaskTemplate.ContainerSpec.Mounts)
}

// The storage screen never sends the artifacts back, since it never shows them:
// a save keeps the service's.
func TestKeepArtifactMounts(t *testing.T) {
	useHostDir(t)
	data := mount.Mount{Type: mount.TypeVolume, Source: "pgdata", Target: "/var/lib/postgresql/data"}
	walg := artifactMount("aa", "/opt/hivepaas/bin/wal-g")
	forged := artifactMount("bb", "/opt/hivepaas/bin/other")

	kept := KeepArtifactMounts([]mount.Mount{data, forged}, []mount.Mount{data, walg})

	assert.Equal(t, []mount.Mount{data, walg}, kept, "an artifact the save names is not the save's to add")
}

func TestFromTemplate(t *testing.T) {
	artifacts := FromTemplate([]*templatemodel.Artifact{{
		Name: "wal-g", Version: "3.0.8", MountPath: "/opt/hivepaas/bin/wal-g",
		Files: map[base.CPUArch]*templatemodel.ArtifactFile{
			base.CPUArchAMD64: {URL: "https://example.com/a", SHA256: "aa", Size: 1},
		},
	}})

	assert.Equal(t, "/opt/hivepaas/bin/wal-g", artifacts.Artifacts[0].MountPath)
	assert.Equal(t, "aa", artifacts.Artifacts[0].Files[base.CPUArchAMD64].SHA256)
	assert.Nil(t, FromTemplate(nil), "a template with none gives the app no setting")
}
