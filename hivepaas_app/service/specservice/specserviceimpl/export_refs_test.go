package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// A project exported alone does not hold the global certificate its app uses.
// The reference becomes an external one: the id finds the certificate again on
// the installation that exported it, the type, name and kind anywhere else.
func TestExportWritesAReferenceOutsideTheExportAsExternal(t *testing.T) {
	path, _ := runExportAt(t, entity.NewObjectScopeProject("p1"), specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": map[string]any{"external": map[string]any{
		"type": "ssl-cert", "name": "localhost", "kind": "self-signed", "id": "cert_1",
	}}}, domain["sslCert"])
}

// The whole installation holds the certificate, so the same reference is a path.
func TestExportWritesAReferenceInsideTheExportAsAPath(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": "global/sslCerts/localhost"}, domain["sslCert"])
}

func exportedStorage(t *testing.T) (*specmodel.Storage, *specmodel.Resources) {
	t.Helper()
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")
	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	deployment := env.Apps["backend"].Deployment
	if !assert.NotNil(t, deployment) || !assert.NotNil(t, deployment.Storage) {
		t.FailNow()
	}
	return deployment.Storage, deployment.Resources
}

// A mount into the app's directory is written the way the storage screen
// writes it: the volume by its path in the bundle, the directory below the
// app's own.
func TestExportWritesAManagedMountAsTheStorageScreenDoes(t *testing.T) {
	storage, _ := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{
		Type: mount.TypeVolume, Source: "projects/project_a/volumes/default",
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "data"},
	}, storage.Mounts["/var/lib/postgresql/data"])
}

// A volume sync discovered is never exported, so a mount into it names the
// volume by an external reference.
func TestExportNamesAVolumeOutsideTheExportByReference(t *testing.T) {
	storage, _ := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{
		Type:          mount.TypeVolume,
		External:      &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared", ID: "gvol_1"},
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "cache"},
	}, storage.Mounts["/shared"])
}

// Every other mount is kept as Docker holds it, and the shared-memory mount is
// carried by resources.memory.shmSize alone.
func TestExportKeepsOtherMountsAsDockerHoldsThem(t *testing.T) {
	storage, resources := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{Type: mount.TypeBind, Source: "/srv/conf"},
		storage.DockerMounts["/etc/app/config"])
	assert.NotContains(t, storage.Mounts, "/dev/shm")
	assert.NotContains(t, storage.DockerMounts, "/dev/shm")
	if assert.NotNil(t, resources) && assert.NotNil(t, resources.Memory) && assert.NotNil(t, resources.Memory.ShmSize) {
		assert.Equal(t, unit.DataSize(64<<20), *resources.Memory.ShmSize)
	}
}
