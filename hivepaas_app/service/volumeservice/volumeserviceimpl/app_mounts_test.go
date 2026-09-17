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
