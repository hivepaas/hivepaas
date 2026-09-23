package volumeserviceimpl

import (
	"encoding/json"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

// An app's data is a directory inside a volume, and the volume holds one of
// those per app. Deleting the app deletes its directory; anything else would
// take the other apps' data with it.
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

// The helper has to see the volume whole: the directory to delete is inside it,
// and mounting with the subpath would put the helper in the directory it is
// meant to remove. The mount it is given must also be its own - writing the
// subpath back into the app's mount is how a clone once emptied the app it was
// copied from.
func TestStorageTargetMountsTheVolumeWhole(t *testing.T) {
	appMount := mount.Mount{
		Type: mount.TypeVolume, Source: "vol-1", Target: "/var/lib/data", ReadOnly: true,
		VolumeOptions: &mount.VolumeOptions{Subpath: "shop/dev/blog"},
	}

	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	target, ok := appStorageTarget(&appMount, volumes)

	assert.True(t, ok)
	assert.Equal(t, "shop/dev/blog", target.subpath)
	assert.Equal(t, volumes[0], target.volume, "the target carries the volume whose directory this is")
	assert.Equal(t, volumeHelperTarget, target.mount.Target)
	assert.False(t, target.mount.ReadOnly, "the helper has to be able to delete")
	assert.Empty(t, target.mount.VolumeOptions.Subpath)
	assert.Equal(t, "shop/dev/blog", appMount.VolumeOptions.Subpath, "the app's own mount is untouched")
}

// Without the volume there is no scope, and without a scope nothing can say
// whose directory this is - so it is nobody's to delete.
func TestStorageTargetRefusesAVolumeTheScopeDoesNotAccountFor(t *testing.T) {
	appMount := mount.Mount{
		Type: mount.TypeVolume, Source: "vol-gone", Target: "/var/lib/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "shop/dev/blog"},
	}

	_, ok := appStorageTarget(&appMount, []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")})

	assert.False(t, ok)
}

// A `local` volume made of a host directory reaches docker as a plain bind, so
// the app's directory has to be found by matching the bind's source against the
// directories the volume settings are made of.
func TestStorageTargetFindsTheAppDirectoryInABind(t *testing.T) {
	volumes := []*entity.Setting{
		clusterVolumeSetting(t, "vol-shallow", "/srv/data"),
		clusterVolumeSetting(t, "vol-deep", "/srv/data/pg"),
	}

	cases := map[string]struct {
		source      string
		wantOK      bool
		wantDevice  string
		wantSubpath string
	}{
		"below a volume's directory":     {"/srv/data/shop/dev/blog", true, "/srv/data", "shop/dev/blog"},
		"below the more specific volume": {"/srv/data/pg/shop/db", true, "/srv/data/pg", "shop/db"},
		"the volume's directory itself":  {"/srv/data", false, "", ""},
		"a path no volume accounts for":  {"/srv/elsewhere/blog", false, "", ""},
		"a path that only looks like it": {"/srv/database/blog", false, "", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			target, ok := appStorageTarget(&mount.Mount{Type: mount.TypeBind, Source: tc.source}, volumes)

			assert.Equal(t, tc.wantOK, ok)
			if !tc.wantOK {
				return
			}
			assert.Equal(t, tc.wantDevice, target.mount.Source, "the helper mounts the volume's own directory")
			assert.Equal(t, tc.wantSubpath, target.subpath)
		})
	}
}

// An unmanaged volume's directory was copied off somebody else's volume rather
// than chosen by HivePaaS, and a bind under it is not ours to delete.
func TestStorageTargetIgnoresABindUnderAnUnmanagedVolume(t *testing.T) {
	vol := clusterVolumeSetting(t, "vol-1", "/srv/data")
	unmanage(t, vol)

	_, ok := appStorageTarget(&mount.Mount{Type: mount.TypeBind, Source: "/srv/data/shop/blog"}, []*entity.Setting{vol})

	assert.False(t, ok)
}

func TestStorageTargetIgnoresMountsWithNothingOfTheAppsInThem(t *testing.T) {
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	for name, mnt := range map[string]mount.Mount{
		"a volume mounted whole": {Type: mount.TypeVolume, Source: "vol-1"},
		"a tmpfs":                {Type: mount.TypeTmpfs, Target: "/tmp"},
		"a named pipe":           {Type: mount.TypeNamedPipe, Source: `\\.\pipe\docker`},
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := appStorageTarget(&mnt, volumes)
			assert.False(t, ok)
		})
	}
}

// Storage pinned to another node cannot be deleted from here: the host path and
// a local container both reach this node's disk, and docker answers a mount of
// a volume it has never heard of by creating an empty one - which would delete
// nothing and report success.
func TestStorageOnAnotherNodeIsNotTouchedLocally(t *testing.T) {
	svc := &service{}

	assert.True(t, svc.storageIsOnThisNode(t.Context(), nil),
		"nothing pinned, so whatever this node can reach is the data")
	assert.True(t, svc.storageIsOnThisNode(t.Context(), []placementservice.VolumePin{{VolumeName: "vol-1"}}))
	assert.False(t, svc.storageIsOnThisNode(t.Context(), []placementservice.VolumePin{
		{VolumeName: "vol-1", NodeLabel: "storage=fast"},
	}), "a node label is not something this node can answer for itself")
}

func TestRemoveAppStorageDoesNothingWithoutMounts(t *testing.T) {
	svc := &service{}

	assert.NoError(t, svc.RemoveAppStorage(t.Context(), nil, nil, nil))
}

func clusterVolumeSetting(t *testing.T, name, device string) *entity.Setting {
	t.Helper()
	data, err := json.Marshal(&entity.ClusterVolume{
		Managed: true,
		Driver:  "local",
		DriverOpts: map[string]string{
			"type":   "none",
			"device": device,
			"o":      "bind,rw",
		},
	})
	assert.NoError(t, err)
	return &entity.Setting{
		ID:    name,
		RefID: name,
		Name:  name,
		Type:  base.SettingTypeClusterVolume,
		Scope: base.ObjectScopeProject,
		Data:  string(data),
	}
}

func unmanage(t *testing.T, setting *entity.Setting) {
	t.Helper()
	vol, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	vol.Managed = false
	setting.Data = string(gofn.Must(json.Marshal(vol)))
}

// The decision the deletion turns on, with nothing else in the way: an app that
// was given the files of the database beside it takes only its own with it.
func TestOwnStorageTargetsLeavesAnotherAppsDirectoryAlone(t *testing.T) {
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}
	app := mountTestApp() // project shop, env prod, app web

	targets := ownStorageTargets(app, []mount.Mount{
		// Its own, as a bind - which is what a managed local volume becomes.
		{Type: mount.TypeBind, Source: "/srv/data/prod/web", Target: "/usr/share/nginx/html"},
		// The database it was put there to look at.
		{Type: mount.TypeBind, Source: "/srv/data/prod/postgres", Target: "/srv/postgres", ReadOnly: true},
		// The same thing as a volume mount, for a volume that stays one.
		{Type: mount.TypeVolume, Source: "vol-1", Target: "/srv/pg2",
			VolumeOptions: &mount.VolumeOptions{Subpath: "prod/postgres"}},
		// Its own again, deeper in.
		{Type: mount.TypeVolume, Source: "vol-1", Target: "/var/cache",
			VolumeOptions: &mount.VolumeOptions{Subpath: "prod/web/cache"}},
	}, volumes)

	subpaths := make([]string, 0, len(targets))
	for _, target := range targets {
		subpaths = append(subpaths, target.subpath)
	}
	assert.Equal(t, []string{"prod/web", "prod/web/cache"}, subpaths)
}

// Nothing to delete is not an error, and neither is a mount this app may see but
// does not own.
func TestOwnStorageTargetsIsEmptyWhenNothingIsTheApps(t *testing.T) {
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	targets := ownStorageTargets(mountTestApp(), []mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/data/prod/postgres", Target: "/srv/postgres"},
		{Type: mount.TypeBind, Source: "/srv/data", Target: "/srv/all"},
		{Type: mount.TypeBind, Source: "/mnt/elsewhere", Target: "/srv/other"},
		{Type: mount.TypeTmpfs, Target: "/tmp"},
	}, volumes)

	assert.Empty(t, targets)
}

// Docker Desktop rewrites a bind source into the path its file-sharing layer
// serves it at, so the spec carries /host_mnt/srv/data where HivePaaS wrote
// /srv/data. Compared literally the mount matches no volume, and the storage the
// operator asked to delete is silently left behind - which is what happened.
func TestStorageTargetFindsTheAppDirectoryBehindADesktopPrefix(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	targets := ownStorageTargets(app, []mount.Mount{{
		Type:   mount.TypeBind,
		Source: "/host_mnt/srv/data/prod/web",
		Target: "/data",
	}}, volumes)

	assert.Len(t, targets, 1)
	assert.Equal(t, "prod/web", targets[0].subpath)
	assert.Equal(t, "/srv/data", targets[0].mount.Source)
}

// Stripping the prefix may only ever produce a path that matches a volume this
// installation configured; a bind that matches nothing still matches nothing.
func TestDesktopPrefixIsNotStrippedIntoSomethingElse(t *testing.T) {
	app := mountTestApp()
	volumes := []*entity.Setting{clusterVolumeSetting(t, "vol-1", "/srv/data")}

	targets := ownStorageTargets(app, []mount.Mount{{
		Type:   mount.TypeBind,
		Source: "/host_mnt/elsewhere/prod/web",
		Target: "/data",
	}}, volumes)

	assert.Empty(t, targets)
}

func TestBindSourceCandidatesKeepsThePathItWasGiven(t *testing.T) {
	assert.Equal(t, []string{"/srv/data/prod/web"}, bindSourceCandidates("/srv/data/prod/web"))
	assert.Equal(t, []string{"/host_mnt/srv/data", "/srv/data"}, bindSourceCandidates("/host_mnt/srv/data"))
	// A directory that merely starts with the same letters is not a prefix.
	assert.Equal(t, []string{"/host_mnts/srv"}, bindSourceCandidates("/host_mnts/srv"))
}
