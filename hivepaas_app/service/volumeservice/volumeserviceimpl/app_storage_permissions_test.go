package volumeserviceimpl

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// The helper is given a command rather than a shell line, so no path can be read
// as one, and it changes directories and files only.
func TestResetPermissionsCmd(t *testing.T) {
	assert.Equal(t,
		[]string{"find", "/mnt/vol/p1/dev/db", "(", "-type", "d", "-o", "-type", "f", ")",
			"-exec", "chown", "999:998", "{}", "+"},
		resetPermissionsCmd("/mnt/vol/p1/dev/db", &volumeservice.StorageOwner{UID: 999, GID: 998}))
	assert.Equal(t,
		[]string{"find", "/mnt/vol/p1/dev/db", "(", "-type", "d", "-o", "-type", "f", ")",
			"-exec", "chmod", "a+rwX", "{}", "+"},
		resetPermissionsCmd("/mnt/vol/p1/dev/db", nil))
}

func TestOpenedModeLetsEveryoneReadAndWriteButMakesNoProgram(t *testing.T) {
	assert.Equal(t, os.FileMode(0o666), openedMode(0o600, false))
	assert.Equal(t, os.FileMode(0o777), openedMode(0o700, true))
	assert.Equal(t, os.FileMode(0o777), openedMode(0o744, false), "a script stays one, for everyone")
}

// Opening a directory up reaches everything the app wrote and nothing it only
// points at: a symlink out of the directory is where another app's data would be.
func TestResetPermissionsOnHostOpensUpAndFollowsNoLink(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "app")
	outside := filepath.Join(root, "other-app", "secret")
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o700))
	assert.NoError(t, os.MkdirAll(filepath.Dir(outside), 0o700))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "data"), nil, 0o600))
	assert.NoError(t, os.WriteFile(outside, nil, 0o600))
	assert.NoError(t, os.Symlink(outside, filepath.Join(dir, "link")))
	assert.NoError(t, syscall.Mkfifo(filepath.Join(dir, "fifo"), 0o600))

	assert.NoError(t, resetPermissionsOnHost(dir, nil))

	assert.Equal(t, os.FileMode(0o777), permOf(t, dir))
	assert.Equal(t, os.FileMode(0o777), permOf(t, filepath.Join(dir, "sub")))
	assert.Equal(t, os.FileMode(0o666), permOf(t, filepath.Join(dir, "sub", "data")))
	assert.Equal(t, os.FileMode(0o600), permOf(t, outside), "the link was not followed")
	assert.Equal(t, os.FileMode(0o600), permOf(t, filepath.Join(dir, "fifo")), "only directories and files")
}

// Giving the files to a user keeps their modes: a 0600 key stays one.
func TestResetPermissionsOnHostGivesFilesToAnOwnerAndKeepsTheirModes(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "key")
	assert.NoError(t, os.WriteFile(key, nil, 0o600))
	// Only root can give a file away; the user running the test can give it to
	// itself, which is enough to see the walk reach it.
	owner := &volumeservice.StorageOwner{UID: os.Getuid(), GID: os.Getgid()}

	assert.NoError(t, resetPermissionsOnHost(dir, owner))

	assert.Equal(t, os.FileMode(0o600), permOf(t, key))
}

// A file that cannot be changed fails the reset, after the others were changed.
func TestResetPermissionsOnHostReportsWhatItCouldNotChange(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can give any file away")
	}
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "data"), nil, 0o600))

	err := resetPermissionsOnHost(dir, &volumeservice.StorageOwner{UID: 0, GID: 0})

	assert.Error(t, err)
}

// A helper that walks what an app wrote is given the app's directory alone, so a
// link out of it has nothing of anybody's to reach.
func TestScopedHelperMountIsTheDirectoryAlone(t *testing.T) {
	vol := &storageTarget{
		mount:   mount.Mount{Type: mount.TypeVolume, Source: "vol-1", Target: volumeHelperTarget},
		subpath: "p1/dev/db",
	}
	bind := &storageTarget{mount: bindMountWhole("/srv/data"), subpath: "p1/dev/db"}

	scopedVol := scopedHelperMount(vol)
	scopedBind := scopedHelperMount(bind)

	assert.Equal(t, "p1/dev/db", scopedVol.VolumeOptions.Subpath)
	assert.Equal(t, "vol-1", scopedVol.Source)
	assert.Nil(t, vol.mount.VolumeOptions, "the target's own mount is left as it was")
	assert.Equal(t, "/srv/data/p1/dev/db", scopedBind.Source)
	assert.Equal(t, volumeHelperTarget, scopedBind.Target)
}
