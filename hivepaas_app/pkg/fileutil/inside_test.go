package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A directory of root: not one out of it, by its path or by a link the
// directory holds, nor a file.
func TestDirInsideStaysInItsRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "apps", "api"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), nil, 0o600))
	assert.NoError(t, os.Symlink(outside, filepath.Join(root, "out")))
	assert.NoError(t, os.Symlink("apps", filepath.Join(root, "apps-link")))

	dir, err := DirInside(root, "apps/api")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "apps", "api"), dir)

	dir, err = DirInside(root, "apps-link/api")
	assert.NoError(t, err, "a link within the root")
	assert.Equal(t, filepath.Join(root, "apps", "api"), dir)

	// A link to root itself is root
	assert.NoError(t, os.Symlink(".", filepath.Join(root, "self")))
	dir, err = DirInside(root, "self")
	assert.NoError(t, err)
	assert.Equal(t, root, dir)

	for _, rel := range []string{"../", "apps/../../x", "/etc", "out", "out/x", "README.md", "missing"} {
		_, err := DirInside(root, rel)
		assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, rel)
	}
}
