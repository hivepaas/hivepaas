package volumeserviceimpl

import (
	"context"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// appMountsSettingRepo answers the two lookups building mounts makes: the
// requested volumes by id, and every volume the app's scope can see.
type appMountsSettingRepo struct {
	repository.SettingRepo
	volumes []*entity.Setting
}

func (f *appMountsSettingRepo) ListByIDs(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, ids []string, _ bool, _ ...bunex.SelectQueryOption,
) ([]*entity.Setting, error) {
	var out []*entity.Setting
	for _, setting := range f.volumes {
		if slices.Contains(ids, setting.ID) {
			out = append(out, setting)
		}
	}
	return out, nil
}

func (f *appMountsSettingRepo) List(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.volumes, nil, nil
}

type recordedHost struct {
	madeSubDirs []string
	permissions int
}

func newAppMountsTest(volumes ...*entity.Setting) (*service, *recordedHost) {
	host := &recordedHost{}
	svc := &service{settingRepo: &appMountsSettingRepo{volumes: volumes}}
	svc.makeSubDirInHost = func(_ context.Context, baseDir, subpath string, _ bool) error {
		host.madeSubDirs = append(host.madeSubDirs, baseDir+"|"+subpath)
		return nil
	}
	svc.ensureVolumePermissions = func(context.Context, *mount.Mount, ...string) error {
		host.permissions++
		return nil
	}
	return svc, host
}

func mountTestApp() *entity.App {
	return &entity.App{
		ID:         "app-1",
		Key:        "web",
		Project:    &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Key: "prod"},
	}
}

func scopedVolume(
	t *testing.T, id, refID string, scope base.ObjectScopeType, vol *entity.ClusterVolume,
) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, RefID: refID, Type: base.SettingTypeClusterVolume, Scope: scope}
	assert.NoError(t, setting.SetData(vol))
	return setting
}

// A bind volume that stays a TypeVolume mount names a volume the target node has
// never heard of, which is the silent-empty-volume failure bind rewriting exists
// to prevent.
func TestBuildAppMountsRewritesABindVolumeIntoABindMount(t *testing.T) {
	svc, host := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp, &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data", "o": "bind"},
	}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-1", Target: "/data"}},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Mounts, 1)
	assert.Equal(t, mount.TypeBind, resp.Mounts[0].Type)
	assert.Equal(t, "/srv/data/web", resp.Mounts[0].Source)
	assert.True(t, resp.Mounts[0].BindOptions.CreateMountpoint)
	assert.Equal(t, []string{"/srv/data|web"}, host.madeSubDirs)
}

// Without the driver config in the spec, the node running the task creates an
// empty default volume under the same name and the app writes to local disk.
func TestBuildAppMountsFillsInTheDriverConfigForANonBindVolume(t *testing.T) {
	svc, host := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject, &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
	}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "files"},
		}},
	})

	assert.NoError(t, err)
	mnt := resp.Mounts[0]
	assert.Equal(t, mount.TypeVolume, mnt.Type)
	assert.Equal(t, "hp-vol-1", mnt.Source)
	assert.Equal(t, "prod/web/files", mnt.VolumeOptions.Subpath, "a project volume is shared, so the app gets env/app")
	assert.Equal(t, "local", mnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", mnt.VolumeOptions.DriverConfig.Options["device"])
	assert.Equal(t, 1, host.permissions)
}

func TestBuildAppMountsKeepsTheKeptMountsFirst(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp, &entity.ClusterVolume{}))
	kept := mount.Mount{Type: mount.TypeVolume, Source: "hp-old", Target: "/old"}

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App:  mountTestApp(),
		Kept: []mount.Mount{kept},
		New:  []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-1", Target: "/new"}},
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"/old", "/new"}, []string{resp.Mounts[0].Target, resp.Mounts[1].Target})
	assert.Equal(t, kept, resp.Mounts[0], "a kept mount is not rebuilt")
}

func TestBuildAppMountsRefusesAnUnknownVolume(t *testing.T) {
	svc, _ := newAppMountsTest()

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-missing", Target: "/data"}},
	})

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestBuildAppMountsRefusesAnotherMountType(t *testing.T) {
	svc, _ := newAppMountsTest()

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeBind, Source: "/srv", Target: "/data"}},
	})

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestBuildAppMountsRefusesMountsPinnedToTwoNodes(t *testing.T) {
	pgdata := newVolumeSetting(t, "vol-pgdata", "ref-pgdata", "pgdata", "node-1")
	uploads := newVolumeSetting(t, "vol-uploads", "ref-uploads", "uploads", "node-2")
	svc, _ := newAppMountsTest(pgdata, uploads)

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App:  mountTestApp(),
		Kept: []mount.Mount{{Type: mount.TypeVolume, Source: "ref-pgdata", Target: "/pg"}},
		New:  []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-uploads", Target: "/uploads"}},
	})

	detail := clientVisibleDetail(t, err)
	assert.Contains(t, detail, "pgdata")
	assert.Contains(t, detail, "uploads")
}

// A mount that names an owner reaches that app's directory, which is the whole
// of what lets a file manager work on the database beside it.
func TestBuildAppMountsPutsAForeignMountInTheOwnersDirectory(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject,
		&entity.ClusterVolume{Managed: true, Driver: "local"}))
	owner := &entity.App{
		ID:         "app-2",
		Key:        "postgres",
		Project:    &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Key: "prod"},
	}

	built, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type:          mount.TypeVolume,
			Source:        "vol-1",
			Target:        "/srv/data",
			ReadOnly:      true,
			OwnerApp:      owner,
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "backups"},
		}},
	})

	assert.NoError(t, err)
	assert.Len(t, built.Mounts, 1)
	assert.Equal(t, "prod/postgres/backups", built.Mounts[0].VolumeOptions.Subpath,
		"the owner's directory, with the request's own subpath below it")
	assert.True(t, built.Mounts[0].ReadOnly)
}

// The same request without an owner is the ordinary case, and must not have
// moved.
func TestBuildAppMountsKeepsTheCallersDirectoryWithoutAnOwner(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-2", "hp-vol-2", base.ObjectScopeProject,
		&entity.ClusterVolume{Managed: true, Driver: "local"}))

	built, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type:          mount.TypeVolume,
			Source:        "vol-2",
			Target:        "/srv/data",
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "backups"},
		}},
	})

	assert.NoError(t, err)
	assert.Equal(t, "prod/web/backups", built.Mounts[0].VolumeOptions.Subpath)
}

// What an app owns is read from the path, so the answer holds for mounts written
// before an app could be given somebody else's directory.
func TestAppOwnsSubpath(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web

	cases := map[string]struct {
		scope   base.ObjectScopeType
		subpath string
		want    bool
	}{
		"its own directory in a project volume":    {base.ObjectScopeProject, "prod/web", true},
		"something below it":                       {base.ObjectScopeProject, "prod/web/uploads", true},
		"another app's directory":                  {base.ObjectScopeProject, "prod/postgres", false},
		"another app whose key starts the same":    {base.ObjectScopeProject, "prod/website", false},
		"another environment":                      {base.ObjectScopeProject, "dev/web", false},
		"the volume root":                          {base.ObjectScopeProject, "", false},
		"its own directory in a global volume":     {base.ObjectScopeGlobal, "shop/prod/web", true},
		"another project in a global volume":       {base.ObjectScopeGlobal, "blog/prod/web", false},
		"its own directory in an app volume":       {base.ObjectScopeApp, "web", true},
		"another app's directory in an app volume": {base.ObjectScopeApp, "postgres", false},
		"a scope that gives no directory":          {base.ObjectScopeHivepaas, "anything", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, appOwnsSubpath(app, tc.scope, tc.subpath))
		})
	}
}

// The deletion path loads the app row without its project and environment, and
// an app whose storage cannot be named is an app whose storage is not deleted -
// never one that deletes by a half-built path.
func TestAppScopePrefixWithoutLoadedRelations(t *testing.T) {
	app := &entity.App{ID: "app-1", Key: "web", ProjectID: "proj-1", ProjectEnvID: "proj-1:prod"}

	assert.Equal(t, "prod/web", appScopePrefix(app, base.ObjectScopeProject),
		"the environment's key is in the id, so a project volume can still be answered for")
	assert.Equal(t, "web", appScopePrefix(app, base.ObjectScopeApp))
	assert.Empty(t, appScopePrefix(app, base.ObjectScopeGlobal),
		"a global volume needs the project's key, which only the relation carries")

	assert.True(t, appOwnsSubpath(app, base.ObjectScopeProject, "prod/web/uploads"))
	assert.False(t, appOwnsSubpath(app, base.ObjectScopeProject, "prod/postgres"))
	assert.False(t, appOwnsSubpath(app, base.ObjectScopeGlobal, "shop/prod/web"),
		"unanswerable is not owned")
}

// An app with no environment at all cannot be told apart from another, so
// nothing in a project volume is its own.
func TestAppScopePrefixWithoutAnEnvironment(t *testing.T) {
	app := &entity.App{ID: "app-1", Key: "web"}

	assert.Empty(t, appScopePrefix(app, base.ObjectScopeProject))
	assert.False(t, appOwnsSubpath(app, base.ObjectScopeProject, "prod/web"))
}

// A volume whose data is the docker socket, or the directory holding it, is
// refused whoever built it and whenever: the Docker API an app may use is
// written down in its Docker API settings, and this would go around them.
func TestBuildAppMountsRefusesAVolumeReachingTheDockerSocket(t *testing.T) {
	for _, device := range []string{"/var/run/docker.sock", "/var/run", "/"} {
		svc, host := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp,
			&entity.ClusterVolume{
				Managed: true, Driver: "local",
				DriverOpts: map[string]string{"type": "none", "device": device, "o": "bind"},
			}))

		_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
			App: mountTestApp(),
			New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-1", Target: "/sock"}},
		})

		assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, device)
		assert.Empty(t, host.madeSubDirs, "%s: nothing is created for a mount that is refused", device)
	}
}

// How a volume is mounted is the volume's own description. A request saying
// otherwise once replaced it, which turned any volume into any mount.
func TestBuildAppMountsTakesTheDriverConfigFromTheVolumeAlone(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp, &entity.ClusterVolume{
		Managed: true, Driver: "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
	}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "files"},
		}},
	})

	assert.NoError(t, err)
	assert.Equal(t, ":/exports/data", resp.Mounts[0].VolumeOptions.DriverConfig.Options["device"])
}

// A subpath is a directory inside the app's own, so it goes below it. One that
// climbs out lands in another app's directory of a shared volume - past the
// permission that mounting another app's storage asks for.
func TestBuildAppMountsRefusesASubpathThatLeavesTheAppsDirectory(t *testing.T) {
	for _, subpath := range []string{"../other", "files/../../other", "/etc", ".."} {
		svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject,
			&entity.ClusterVolume{Managed: true, Driver: "local"}))

		_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
			App: mountTestApp(),
			New: []*volumeservice.AppMountReq{{
				Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
				VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: subpath},
			}},
		})

		assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, subpath)
	}
}

// What a subpath is for still works: a directory below the app's own.
func TestBuildAppMountsTakesASubpathBelowTheAppsDirectory(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject,
		&entity.ClusterVolume{Managed: true, Driver: "local"}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "files/uploads"},
		}},
	})

	assert.NoError(t, err)
	assert.Equal(t, "prod/web/files/uploads", resp.Mounts[0].VolumeOptions.Subpath)
}
