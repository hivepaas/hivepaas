package volumeserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The screen has to be able to say "the files of postgres" where the service
// spec says only a path.
func TestDescribeAppMountNamesTheOwner(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web
	volumes := []*entity.Setting{
		scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject, &entity.ClusterVolume{Managed: true}),
	}

	cases := map[string]struct {
		subpath string
		wantKey string
		wantOwn bool
		wantSub string
	}{
		"its own directory":             {"prod/web", "web", true, ""},
		"below its own directory":       {"prod/web/uploads", "web", true, "uploads"},
		"another app's directory":       {"prod/postgres", "postgres", false, ""},
		"below another app's directory": {"prod/postgres/pgdata", "postgres", false, "pgdata"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mnt := &mount.Mount{
				Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/data",
				VolumeOptions: &mount.VolumeOptions{Subpath: tc.subpath},
			}

			desc := describeAppMount(app, mnt, volumes)

			assert.Equal(t, tc.wantKey, desc.AppKey)
			assert.Equal(t, tc.wantOwn, desc.Own)
			assert.Equal(t, tc.wantSub, desc.Subpath)
			assert.Equal(t, "vol-1", desc.VolumeID)
		})
	}
}

// A managed local volume reaches docker as a bind, and the app that owns the
// directory has to be readable out of the host path just the same.
func TestDescribeAppMountReadsABind(t *testing.T) {
	app := mountTestApp()
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	desc := describeAppMount(app, &mount.Mount{
		Type: mount.TypeBind, Source: "/srv/data/prod/postgres", Target: "/data",
	}, volumes)

	assert.Equal(t, "postgres", desc.AppKey)
	assert.False(t, desc.Own)
	assert.Equal(t, "vol-1", desc.VolumeID)
}

// A volume mounted whole, or one this scope knows nothing about, is nobody's
// directory - and saying so is what keeps the screen from inventing an owner.
func TestDescribeAppMountLeavesTheUnknownUnnamed(t *testing.T) {
	app := mountTestApp()
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	cases := map[string]mount.Mount{
		"the volume root":             {Type: mount.TypeBind, Source: "/srv/data", Target: "/data"},
		"a bind nothing accounts for": {Type: mount.TypeBind, Source: "/mnt/elsewhere", Target: "/data"},
		"a volume with no subpath":    {Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/data"},
	}
	for name, mnt := range cases {
		t.Run(name, func(t *testing.T) {
			desc := describeAppMount(app, &mnt, volumes)
			assert.Empty(t, desc.AppKey)
			assert.False(t, desc.Own)
			assert.Empty(t, desc.VolumeID)
		})
	}
}

// An app is found again by its key only inside its own environment - the same
// key in another environment is another app. A directory in a volume shared
// wider than one environment is named only when it lies in this app's.
func TestDescribeAppMountNamesNoAppOutsideItsEnvironment(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web
	projectVolume := scopedVolume(t, "vol-p", "hp-vol-p", base.ObjectScopeProject, &entity.ClusterVolume{Managed: true})
	globalVolume := scopedVolume(t, "vol-g", "hp-vol-g", base.ObjectScopeGlobal, &entity.ClusterVolume{Managed: true})

	cases := map[string]struct {
		source  string
		subpath string
		wantKey string
	}{
		"a project volume, this environment":    {"hp-vol-p", "prod/postgres", "postgres"},
		"a project volume, another environment": {"hp-vol-p", "staging/postgres", ""},
		"a global volume, this environment":     {"hp-vol-g", "shop/prod/postgres", "postgres"},
		"a global volume, another environment":  {"hp-vol-g", "shop/staging/postgres", ""},
		"a global volume, another project":      {"hp-vol-g", "blog/prod/postgres", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			desc := describeAppMount(app, &mount.Mount{
				Type: mount.TypeVolume, Source: tc.source, Target: "/data",
				VolumeOptions: &mount.VolumeOptions{Subpath: tc.subpath},
			}, []*entity.Setting{projectVolume, globalVolume})

			assert.Equal(t, tc.wantKey, desc.AppKey)
			if tc.wantKey == "" {
				assert.Empty(t, desc.VolumeID, "a directory nobody is named for names no volume either")
			}
		})
	}
}
