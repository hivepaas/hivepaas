package volumeserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// The check has to look at the directory the mount will be made at. Looking
// anywhere else reports on somebody else's data, and reports the app's own as
// clean - which is the failure this whole feature exists to prevent.
func TestAppStoragePathIsWhereTheMountWillBe(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web

	for _, tc := range []struct {
		scope base.ObjectScopeType
		want  string
	}{
		{base.ObjectScopeGlobal, "shop/prod/web"},
		{base.ObjectScopeProject, "prod/web"},
		{base.ObjectScopeProjectEnv, "web"},
		{base.ObjectScopeApp, "web"},
	} {
		assert.Equal(t, tc.want, appStoragePath(app, tc.scope, ""), "scope %s", tc.scope)

		// And it agrees with what the mount builder would have written.
		setting := clusterVolumeSetting(t, "vol-1", "/srv/data")
		setting.Scope = tc.scope
		built := calcMountSubpath(app, &volumeservice.AppMountReq{Type: mount.TypeVolume}, setting)
		assert.Equal(t, built, appStoragePath(app, tc.scope, ""), "scope %s", tc.scope)
	}
}

func TestAppStoragePathTakesTheMountsOwnSubpath(t *testing.T) {
	app := mountTestApp()
	assert.Equal(t, "prod/web/uploads", appStoragePath(app, base.ObjectScopeProject, "uploads"))
	// Anything that is not a plain relative path is refused rather than guessed
	// at: the answer is used to read a directory, and above it is not the app's.
	assert.Empty(t, appStoragePath(app, base.ObjectScopeProject, "../../etc"))
}

func TestAppStoragePathNeedsTheKeysItIsBuiltFrom(t *testing.T) {
	assert.Empty(t, appStoragePath(&entity.App{Key: "web"}, base.ObjectScopeProject, ""))
	assert.Empty(t, appStoragePath(
		&entity.App{Key: "web", ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
		base.ObjectScopeGlobal, ""))
}

// A state nothing could be read for must not read as "there is nothing there".
func TestStorageStateSaysNothingWhenItCouldNotLook(t *testing.T) {
	assert.False(t, (&volumeservice.AppStorageState{Exists: true, Empty: false}).HasData())
	assert.True(t, (&volumeservice.AppStorageState{Checked: true, Exists: true}).HasData())
	assert.False(t, (&volumeservice.AppStorageState{Checked: true, Exists: true, Empty: true}).HasData())
}
