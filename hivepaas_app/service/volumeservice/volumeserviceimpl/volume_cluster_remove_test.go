package volumeserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Removing a volume made of a host directory leaves every byte of it on disk,
// so that directory is deleted separately - but only one HivePaaS chose, which
// means one under the configured storage root. Anywhere else the operator picked
// the directory, and it can be a mount point or hold things that were never
// ours.
func TestVolumeStorageTargetIsOnlyADirectoryHivepaasChose(t *testing.T) {
	withStorageRoot(t, "/srv/hivepaas")

	cases := map[string]struct {
		vol         *entity.ClusterVolume
		wantOK      bool
		wantParent  string
		wantSubpath string
	}{
		"a directory under the storage root": {
			bindVolume("/srv/hivepaas/project_data/shop"),
			true, "/srv/hivepaas/project_data", "shop",
		},
		"a directory deeper under the storage root": {
			bindVolume("/srv/hivepaas/project_data/shop/dev"),
			true, "/srv/hivepaas/project_data/shop", "dev",
		},
		"a trailing slash": {
			bindVolume("/srv/hivepaas/project_data/shop/"),
			true, "/srv/hivepaas/project_data", "shop",
		},
		"the storage root itself":        {bindVolume("/srv/hivepaas"), false, "", ""},
		"a directory the operator chose": {bindVolume("/mnt/nas/shop"), false, "", ""},
		"a path that only looks like it": {bindVolume("/srv/hivepaas-old/shop"), false, "", ""},
		"a directory that climbs out":    {bindVolume("/srv/hivepaas/project_data/.."), false, "", ""},
		"a plain local volume": {
			&entity.ClusterVolume{Managed: true, Driver: "local"}, false, "", "",
		},
		"a volume discovered from docker": {unmanagedBindVolume("/srv/hivepaas/project_data/shop"), false, "", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			target, ok := volumeStorageTarget(tc.vol)

			assert.Equal(t, tc.wantOK, ok)
			if !tc.wantOK {
				return
			}
			assert.Equal(t, tc.wantParent, target.mount.Source,
				"the helper mounts the directory above, so the directory itself can go")
			assert.Equal(t, tc.wantSubpath, target.subpath)
		})
	}
}

// Without a storage root there is no directory HivePaaS can claim to have
// chosen, so nothing is deleted.
func TestVolumeStorageTargetNeedsAStorageRoot(t *testing.T) {
	withStorageRoot(t, "")

	_, ok := volumeStorageTarget(bindVolume("/srv/hivepaas/project_data/shop"))

	assert.False(t, ok)
}

// A storage root of "/" would make every directory on the node one HivePaaS
// chose.
func TestVolumeStorageTargetRefusesTheWholeFilesystemAsARoot(t *testing.T) {
	withStorageRoot(t, "/")

	_, ok := volumeStorageTarget(bindVolume("/etc"))

	assert.False(t, ok)
}

func withStorageRoot(t *testing.T, root string) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{BindSource: root}})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func bindVolume(device string) *entity.ClusterVolume {
	return &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": device, "o": "bind,rw"},
	}
}

func unmanagedBindVolume(device string) *entity.ClusterVolume {
	vol := bindVolume(device)
	vol.Managed = false
	return vol
}
