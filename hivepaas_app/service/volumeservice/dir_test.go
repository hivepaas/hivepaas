package volumeservice

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runSh(t *testing.T, cmd string) {
	t.Helper()
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if !assert.NoError(t, err, string(out)) {
		t.FailNow()
	}
}

func modeOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return info.Mode().Perm()
}

// A directory the helper has just made is opened up, or an app that does not
// run as root could not write to its own storage.
func TestMakeDirWritableCmdOpensANewDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "app", "data dir")

	runSh(t, MakeDirWritableCmd(dir))

	assert.Equal(t, os.FileMode(0o777), modeOf(t, dir))
}

// Once the app has written to it, the directory and everything in it keep the
// modes the app gave them.
func TestMakeDirWritableCmdLeavesAnAppsDataAlone(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pgdata")
	assert.NoError(t, os.Mkdir(dir, 0o700))
	key := filepath.Join(dir, "secret-key.txt")
	assert.NoError(t, os.WriteFile(key, []byte("k"), 0o600))
	assert.NoError(t, os.Chmod(dir, 0o700))

	runSh(t, MakeDirWritableCmd(dir))

	assert.Equal(t, os.FileMode(0o700), modeOf(t, dir))
	assert.Equal(t, os.FileMode(0o600), modeOf(t, key))
}

// A path is handed to the shell quoted, whatever it holds.
func TestMakeDirWritableCmdQuotesThePath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "it's $(touch pwned)")

	runSh(t, MakeDirWritableCmd(dir))

	assert.DirExists(t, dir)
	assert.NoFileExists(t, "pwned")
}
