package volumeserviceimpl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func permOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return info.Mode().Perm()
}

// A directory reached on the host follows the same rule as one reached through
// a helper container: opened up while empty, left alone once the app has used it.
func TestEnsurePermissionsOnDirectHostPath(t *testing.T) {
	base := filepath.Join(t.TempDir(), "vol")
	used := filepath.Join(base, "p1", "dev", "db")
	assert.NoError(t, os.MkdirAll(used, 0o700))
	assert.NoError(t, os.WriteFile(filepath.Join(used, "PG_VERSION"), []byte("18"), 0o600))
	assert.NoError(t, os.Chmod(used, 0o700))

	err := (&service{}).ensurePermissionsOnDirectHostPath(base, "p1/dev/db", "p1/dev/web")

	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), permOf(t, filepath.Join(base, "p1", "dev", "web")), "new, so opened up")
	assert.Equal(t, os.FileMode(0o700), permOf(t, used), "in use, so left as the app set it")
	assert.Equal(t, os.FileMode(0o600), permOf(t, filepath.Join(used, "PG_VERSION")))
	assert.NotEqual(t, os.FileMode(0o777), permOf(t, base), "the root holds apps' directories")
}
