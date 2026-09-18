package appcloneserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
)

func volumeCloneData() *appCloneData {
	return &appCloneData{
		AppCloneReq: &appcloneservice.AppCloneReq{SrcApp: &entity.App{ID: "src-1", Key: "blog"}},
		DestApp:     &entity.App{ID: "dst-1", Key: "blog_copy"},
	}
}

// A mount's options live behind a pointer, and the copy starts as a shallow copy
// of the source's mount. Writing the copy's subpath through that pointer moved
// the source app's data directory to a path named after the copy: the source
// then read an empty directory, and its data was still where it had always been
// with nothing pointing at it.
func TestCalcVolumeMountSubpathLeavesTheSourceMountAlone(t *testing.T) {
	svc := &service{}
	srcMount := mount.Mount{
		Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "blog"},
	}
	destMount := srcMount

	assert.True(t, svc.calcVolumeMountSubpath(&destMount, &srcMount, volumeCloneData()))

	assert.Equal(t, "blog_copy", destMount.VolumeOptions.Subpath)
	assert.Equal(t, "blog", srcMount.VolumeOptions.Subpath, "the source app still reads its own directory")
	assert.NotSame(t, srcMount.VolumeOptions, destMount.VolumeOptions)
}

// Cloning the same app twice has to give the second copy a volume too, which it
// did not while the first clone had rewritten the source's subpath: the name to
// replace was gone, so the subpath came back unchanged and the volume was
// dropped without a word.
func TestCalcVolumeMountSubpathWorksTheSecondTime(t *testing.T) {
	svc := &service{}
	srcMount := mount.Mount{
		Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "blog"},
	}

	first := srcMount
	assert.True(t, svc.calcVolumeMountSubpath(&first, &srcMount, volumeCloneData()))

	second := srcMount
	data := volumeCloneData()
	data.DestApp = &entity.App{ID: "dst-2", Key: "blog_copy2"}
	assert.True(t, svc.calcVolumeMountSubpath(&second, &srcMount, data))
	assert.Equal(t, "blog_copy2", second.VolumeOptions.Subpath)
}

// A volume the source app has no subpath in is shared as it is, and there is
// nothing to rename.
func TestCalcVolumeMountSubpathSkipsAVolumeWithoutOne(t *testing.T) {
	svc := &service{}
	srcMount := mount.Mount{Type: mount.TypeVolume, Source: "vol-1", Target: "/data"}
	destMount := srcMount

	assert.False(t, svc.calcVolumeMountSubpath(&destMount, &srcMount, volumeCloneData()))
}
