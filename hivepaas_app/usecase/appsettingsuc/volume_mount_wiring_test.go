package appsettingsuc

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// stubVolumeService answers only what building a mount actually calls. The
// embedded nil interface turns any other method into a panic, so a test that
// starts reaching further fails loudly instead of quietly doing nothing.
type stubVolumeService struct {
	volumeservice.Service

	madeSubDirs []string
	makeErr     error
}

func (s *stubVolumeService) MakeSubDirInHost(
	_ context.Context,
	baseDirInHost string,
	subpath string,
	_ bool,
) error {
	s.madeSubDirs = append(s.madeSubDirs, baseDirInHost+"|"+subpath)
	return s.makeErr
}

func (s *stubVolumeService) EnsureVolumePermissions(
	_ context.Context,
	_ *mount.Mount,
	_ ...string,
) error {
	return nil
}

// newStorageSettingsData is the smallest data an app storage update carries
// that buildDockerMount can be driven from: the app supplies the scope-derived
// subpath and nothing else is read.
func newStorageSettingsData(appKey string) *updateAppStorageSettingsData {
	return &updateAppStorageSettingsData{App: &entity.App{Key: appKey}}
}

func appScopedVolumeSetting(refID string) *entity.Setting {
	return &entity.Setting{RefID: refID, Scope: base.ObjectScopeApp}
}

// buildDockerMount has to call useBindMountIfAppropriate: a bind volume that
// stays a TypeVolume mount names a volume the target node has never heard of,
// which is the silent-empty-volume failure this branch exists to remove. The
// leaf helper is covered on its own; this covers that it is reached.
func TestBuildDockerMountRewritesABindVolumeIntoABindMount(t *testing.T) {
	volumeService := &stubVolumeService{}
	uc := &UC{volumeService: volumeService}

	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "01JVOL", Target: "/data"}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data", "o": "bind"},
	}

	uc.buildDockerMount(context.Background(), dockerMnt, &appsettingsdto.Mount{Type: mount.TypeVolume},
		vol, appScopedVolumeSetting("01JVOL"), newStorageSettingsData("web"))

	assert.Equal(t, mount.TypeBind, dockerMnt.Type)
	assert.Equal(t, "/srv/data/web", dockerMnt.Source)
	if dockerMnt.BindOptions == nil {
		t.Fatal("expected BindOptions to be set")
	}
	assert.True(t, dockerMnt.BindOptions.CreateMountpoint)
	assert.Equal(t, []string{"/srv/data|web"}, volumeService.madeSubDirs)
}

// And it has to call applyVolumeDriverConfigUnlessOverridden for everything that
// stays a volume mount: without the driver config in the spec, the node running
// the task creates an empty default volume under the same name and the app
// writes to local disk instead of the nfs export.
func TestBuildDockerMountFillsInTheDriverConfigForANonBindVolume(t *testing.T) {
	uc := &UC{volumeService: &stubVolumeService{}}

	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "01JVOL", Target: "/data"}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
	}

	uc.buildDockerMount(context.Background(), dockerMnt, &appsettingsdto.Mount{Type: mount.TypeVolume},
		vol, appScopedVolumeSetting("01JVOL"), newStorageSettingsData("web"))

	assert.Equal(t, mount.TypeVolume, dockerMnt.Type)
	if dockerMnt.VolumeOptions == nil || dockerMnt.VolumeOptions.DriverConfig == nil {
		t.Fatalf("expected VolumeOptions.DriverConfig to be set, got %+v", dockerMnt.VolumeOptions)
	}
	assert.Equal(t, "local", dockerMnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", dockerMnt.VolumeOptions.DriverConfig.Options["device"])
}
