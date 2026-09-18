package volumeserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
)

// An app's data is a directory inside a volume, named after the app, and the
// volume holds one of those per app. Deleting the app deletes its directory;
// anything else would take the other apps' data with it.
func TestMountSubpathIsWhatAnAppOwnsInAVolume(t *testing.T) {
	cases := map[string]struct {
		mnt  mount.Mount
		want string
	}{
		"a volume with a subpath": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "blog"}},
			"blog",
		},
		"a subpath with slashes around it": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "/blog/data/"}},
			"blog/data",
		},
		"a cluster volume": {
			mount.Mount{Type: mount.TypeCluster, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "blog"}},
			"blog",
		},
		"a volume mounted whole": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1"},
			"",
		},
		"a volume whose subpath is empty": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1", VolumeOptions: &mount.VolumeOptions{}},
			"",
		},
		"a subpath that climbs out of the volume": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "blog/../../etc"}},
			"",
		},
		"a subpath that is nothing but a traversal": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: ".."}},
			"",
		},
		"a subpath carrying a shell command": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "blog'; rm -rf /; echo '"}},
			"",
		},
		"a subpath with an empty segment": {
			mount.Mount{Type: mount.TypeVolume, Source: "vol-1",
				VolumeOptions: &mount.VolumeOptions{Subpath: "blog//data"}},
			"",
		},
		"a bind mount": {
			mount.Mount{Type: mount.TypeBind, Source: "/srv/data"},
			"",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, mountSubpath(&tc.mnt))
		})
	}
}

// Nothing is removed for a mount with no usable subpath of its own. That is the
// whole volume, which belongs to whoever created it and may be shared; a bind
// mount is a path on the host that HivePaaS did not make; and a subpath that
// points outside the volume is not this app's to delete.
func TestRemoveAppStorageLeavesAWholeVolumeAlone(t *testing.T) {
	svc := &service{}

	err := svc.RemoveAppStorage(t.Context(), []mount.Mount{
		{Type: mount.TypeVolume, Source: "vol-1"},
		{Type: mount.TypeVolume, Source: "vol-2", VolumeOptions: &mount.VolumeOptions{Subpath: "../elsewhere"}},
		{Type: mount.TypeBind, Source: "/srv/data"},
		{Type: mount.TypeTmpfs, Target: "/tmp"},
	})

	assert.NoError(t, err, "nothing to remove, and nothing to reach docker for")
}

func TestRemoveAppStorageDoesNothingWithoutMounts(t *testing.T) {
	svc := &service{}

	assert.NoError(t, svc.RemoveAppStorage(t.Context(), nil))
}
