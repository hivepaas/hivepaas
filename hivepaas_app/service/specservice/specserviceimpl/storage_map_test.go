package specserviceimpl

import (
	"maps"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func storageTask(mounts ...mount.Mount) *swarm.TaskSpec {
	return &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Mounts: mounts}}
}

func fixedVolumeRef(paths map[string]string, external map[string]*specmodel.ExternalRef) volumeRef {
	return func(volumeID string) (string, *specmodel.ExternalRef) {
		return paths[volumeID], external[volumeID]
	}
}

// A mount into the app's own directory is written the way the storage screen
// writes it, whatever form Docker was handed: a managed local volume reaches
// Docker as a bind.
func TestMapAppStorageWritesTheAppsOwnDirectoryAsAVolume(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeBind, Source: "/srv/data/prod/web/uploads", Target: "/data", ReadOnly: true,
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, Subpath: "uploads", VolumeID: "vol-1"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(map[string]string{"vol-1": "projects/shop/volumes/data"}, nil))

	assert.NoError(t, err)
	assert.Equal(t, map[string]specmodel.Mount{
		"/data": {
			Type: mount.TypeVolume, Source: "projects/shop/volumes/data", ReadOnly: true,
			VolumeOptions: &specmodel.VolumeOptions{Subpath: "uploads"},
		},
	}, out.Mounts)
	assert.Empty(t, out.DockerMounts)
}

// Another app's directory names that app, and whether it may be changed is what
// Write says - the builder derives the mount's read-only flag from it.
func TestMapAppStorageNamesAnotherAppsDirectory(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/pg", ReadOnly: true,
		VolumeOptions: &mount.VolumeOptions{Subpath: "prod/postgres/pgdata", NoCopy: true},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "postgres", Subpath: "pgdata", VolumeID: "vol-1"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(map[string]string{"vol-1": "projects/shop/volumes/data"}, nil))

	assert.NoError(t, err)
	assert.Equal(t, specmodel.Mount{
		Type: mount.TypeVolume, Source: "projects/shop/volumes/data",
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "pgdata", NoCopy: true},
		SourceApp:     &specmodel.MountSourceApp{App: "postgres", Write: false},
	}, out.Mounts["/pg"])
}

// A volume the export does not hold is named by an external reference.
func TestMapAppStorageNamesAVolumeOutsideTheExport(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-shared", Target: "/shared",
		VolumeOptions: &mount.VolumeOptions{Subpath: "shop/prod/web"},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, VolumeID: "gvol-1"}}
	shared := &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared", ID: "gvol-1"}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, map[string]*specmodel.ExternalRef{"gvol-1": shared}))

	assert.NoError(t, err)
	assert.Equal(t, specmodel.Mount{Type: mount.TypeVolume, External: shared}, out.Mounts["/shared"])
}

// Every other mount is kept as Docker holds it, and the shared-memory mount is
// kept by neither map: resources.memory.shmSize carries it.
func TestMapAppStorageKeepsOtherMountsAsDockerHoldsThem(t *testing.T) {
	task := storageTask(
		mount.Mount{Type: mount.TypeTmpfs, Target: "/dev/shm", TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20}},
		mount.Mount{Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime", ReadOnly: true},
		mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/whole"},
	)
	descs := []*volumeservice.AppMountDesc{{}, {}, {}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, nil))

	assert.NoError(t, err)
	assert.Empty(t, out.Mounts)
	assert.Equal(t, []string{"/etc/localtime", "/whole"}, slices.Sorted(maps.Keys(out.DockerMounts)))
	assert.Equal(t, specmodel.Mount{Type: mount.TypeBind, Source: "/etc/localtime", ReadOnly: true},
		out.DockerMounts["/etc/localtime"])
}

// A volume nothing can name any more - its setting is gone - leaves the mount as
// Docker holds it rather than losing it.
func TestMapAppStorageKeepsAMountWhoseVolumeCannotBeNamed(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-gone", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "prod/web"},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, VolumeID: "gone"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, nil))

	assert.NoError(t, err)
	assert.Empty(t, out.Mounts)
	assert.Equal(t, "hp-gone", out.DockerMounts["/data"].Source)
}

// Nothing upstream validates that two mounts do not share a target, so this is
// the only thing holding the invariant a future snapshot pins data to.
func TestMapAppStorageRefusesDuplicateMountTargets(t *testing.T) {
	task := storageTask(
		mount.Mount{Type: mount.TypeVolume, Source: "vol_1", Target: "/data"},
		mount.Mount{Type: mount.TypeBind, Source: "/srv", Target: "/data"},
	)

	_, err := mapAppStorage(task, []*volumeservice.AppMountDesc{{}, {}}, fixedVolumeRef(nil, nil))

	assert.ErrorIs(t, err, hperrors.ErrSpecMountTargetDuplicated)
}

func TestMapAppStorageIsNilWithoutMounts(t *testing.T) {
	out, err := mapAppStorage(storageTask(), nil, fixedVolumeRef(nil, nil))
	assert.NoError(t, err)
	assert.Nil(t, out)

	out, err = mapAppStorage(&swarm.TaskSpec{}, nil, fixedVolumeRef(nil, nil))
	assert.NoError(t, err)
	assert.Nil(t, out)
}
