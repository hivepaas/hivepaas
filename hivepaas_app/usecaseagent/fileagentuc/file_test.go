package fileagentuc

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// newHostUC is the use case over a fake host: host/ stands for the node's /,
// and the volume's directory is /srv/vol on it.
func newHostUC(t *testing.T) (*UC, string) {
	t.Helper()
	host := t.TempDir()
	if err := os.MkdirAll(filepath.Join(host, "srv", "vol"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &UC{hostPrefix: host}, host
}

const root = "/srv/vol"

func TestAFileIsWrittenReadStatedAndRemovedInsideTheRoot(t *testing.T) {
	uc, host := newHostUC(t)
	ctx := context.Background()

	size, err := uc.Write(ctx, root, ".hivepaas/cache/repos/r1.tar.lz4", strings.NewReader("cache bytes"))
	assert.NoError(t, err)
	assert.Equal(t, int64(len("cache bytes")), size)

	info, err := os.Stat(filepath.Join(host, "srv/vol/.hivepaas"))
	if assert.NoError(t, err) {
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(), "HivePaaS's own directory")
	}

	stated, err := uc.Stat(root, ".hivepaas/cache/repos/r1.tar.lz4")
	assert.NoError(t, err)
	assert.Equal(t, size, stated)

	reader, readSize, err := uc.Open(root, ".hivepaas/cache/repos/r1.tar.lz4")
	if assert.NoError(t, err) {
		content, _ := io.ReadAll(reader)
		_ = reader.Close()
		assert.Equal(t, "cache bytes", string(content))
		assert.Equal(t, size, readSize)
	}

	assert.NoError(t, uc.Remove(root, ".hivepaas/cache/repos/r1.tar.lz4"))
	_, err = uc.Stat(root, ".hivepaas/cache/repos/r1.tar.lz4")
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	assert.NoError(t, uc.Remove(root, ".hivepaas/cache/repos/r1.tar.lz4"), "removing what is gone is done")
}

// Nothing a caller passes reaches outside the volume's directory.
func TestAPathLeavingTheRootIsRefused(t *testing.T) {
	uc, host := newHostUC(t)
	if err := os.WriteFile(filepath.Join(host, "etc-passwd"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(host, filepath.Join(host, "srv/vol/escape")); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"../../etc-passwd", "/etc-passwd", "a/../../x", "", ".", "escape/etc-passwd"} {
		_, err := uc.Stat(root, path)
		assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, "stat %q", path)
		_, _, err = uc.Open(root, path)
		assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, "open %q", path)
		_, err = uc.Write(context.Background(), root, path, strings.NewReader("x"))
		assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, "write %q", path)
		assert.ErrorIs(t, uc.Remove(root, path), hperrors.ErrFilePathOutsideRoot, "remove %q", path)
	}
	_, err := uc.Stat("/", "etc-passwd")
	assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, "the host's root is no volume")
	_, err = uc.Stat("srv/vol", "x")
	assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, "a relative root")
}

type failingReader struct{ sent bool }

func (r *failingReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		return copy(p, "half"), nil
	}
	return 0, errors.New("connection lost")
}

// A write that fails half way leaves nothing: neither the file nor its
// temporary.
func TestAFailedWriteLeavesNothing(t *testing.T) {
	uc, host := newHostUC(t)

	_, err := uc.Write(context.Background(), root, "out/f.bin", &failingReader{})

	assert.Error(t, err)
	entries, _ := os.ReadDir(filepath.Join(host, "srv/vol/out"))
	assert.Empty(t, entries)
}

// A write replaces what was there only once it has all of it.
func TestAWriteReplacesTheFileWhole(t *testing.T) {
	uc, host := newHostUC(t)
	ctx := context.Background()
	_, err := uc.Write(ctx, root, "f.txt", strings.NewReader("old"))
	assert.NoError(t, err)

	_, err = uc.Write(ctx, root, "f.txt", &failingReader{})
	assert.Error(t, err)
	content, _ := os.ReadFile(filepath.Join(host, "srv/vol/f.txt"))
	assert.Equal(t, "old", string(content))

	_, err = uc.Write(ctx, root, "f.txt", strings.NewReader("new"))
	assert.NoError(t, err)
	content, _ = os.ReadFile(filepath.Join(host, "srv/vol/f.txt"))
	assert.Equal(t, "new", string(content))
}
