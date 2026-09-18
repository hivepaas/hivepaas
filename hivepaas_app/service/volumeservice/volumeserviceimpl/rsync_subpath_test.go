package volumeserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
)

// The helper container has to see each volume whole: the rsync command reaches
// into the subpath itself, and creates the destination one when it is not there.
// Mounting with the subpath as well applied it twice - docker mounted inside it
// and the command looked for it again in there, so the source never existed and
// rsync stopped without copying anything.
func TestMountWholeVolumeDropsTheSubpath(t *testing.T) {
	src := &mount.Mount{
		Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{
			Subpath: "blog",
			Labels:  map[string]string{"hivepaas.volume.name": "data"},
		},
	}

	out := mountWholeVolume(src)

	assert.Equal(t, "vol-1", out.Source)
	assert.Empty(t, out.VolumeOptions.Subpath)
	assert.Equal(t, map[string]string{"hivepaas.volume.name": "data"}, out.VolumeOptions.Labels,
		"everything else about the volume is kept")
}

// The mount handed in belongs to a running app, so clearing the subpath must not
// reach it.
func TestMountWholeVolumeLeavesTheCallersMountAlone(t *testing.T) {
	src := &mount.Mount{
		Type: mount.TypeVolume, Source: "vol-1",
		VolumeOptions: &mount.VolumeOptions{Subpath: "blog"},
	}

	out := mountWholeVolume(src)

	assert.Equal(t, "blog", src.VolumeOptions.Subpath)
	assert.NotSame(t, src.VolumeOptions, out.VolumeOptions)
}

func TestMountWholeVolumeHandlesAMountWithoutOptions(t *testing.T) {
	src := &mount.Mount{Type: mount.TypeBind, Source: "/srv/data"}

	out := mountWholeVolume(src)

	assert.Equal(t, "/srv/data", out.Source)
	assert.Nil(t, out.VolumeOptions)
}
