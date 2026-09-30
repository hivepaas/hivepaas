package srcpack

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// A source tree arrives as it left: nested directories, an executable file, an
// empty directory and a symbolic link.
func TestASourceTreeArrivesAsItLeft(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	must(t, os.MkdirAll(filepath.Join(src, "cmd", "app"), 0o755))
	must(t, os.MkdirAll(filepath.Join(src, "empty"), 0o755))
	must(t, os.WriteFile(filepath.Join(src, "Dockerfile"), []byte("FROM scratch\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(src, "cmd", "app", "main.go"), []byte("package main\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(src, "run.sh"), []byte("#!/bin/sh\n"), 0o755))
	must(t, os.Symlink("cmd/app/main.go", filepath.Join(src, "main-link")))

	var packed bytes.Buffer
	sent, err := Pack(context.Background(), src, &packed)
	must(t, err)
	got, err := Unpack(context.Background(), &packed, dst)
	must(t, err)

	assert.Equal(t, 3, sent.Files)
	assert.Equal(t, int64(len("FROM scratch\n")+len("package main\n")+len("#!/bin/sh\n")), sent.Bytes)
	assert.Equal(t, sent, got)

	content, err := os.ReadFile(filepath.Join(dst, "cmd", "app", "main.go"))
	must(t, err)
	assert.Equal(t, "package main\n", string(content))
	info, err := os.Stat(filepath.Join(dst, "run.sh"))
	must(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	assert.DirExists(t, filepath.Join(dst, "empty"))
	target, err := os.Readlink(filepath.Join(dst, "main-link"))
	must(t, err)
	assert.Equal(t, "cmd/app/main.go", target)
}

type entry struct {
	header tar.Header
	body   string
}

// archive is a packed source holding exactly these entries, as a hostile or
// broken sender could make it.
func archive(t *testing.T, entries ...entry) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	must(t, err)
	tw := tar.NewWriter(zw)
	for _, e := range entries {
		h := e.header
		h.Size = int64(len(e.body))
		if h.Mode == 0 {
			h.Mode = 0o644
		}
		must(t, tw.WriteHeader(&h))
		_, err = tw.Write([]byte(e.body))
		must(t, err)
	}
	must(t, tw.Close())
	must(t, zw.Close())
	return &buf
}

// An entry that would be written outside the directory is refused, and nothing
// is written there.
func TestAnEntryLeavingTheDirectoryIsRefused(t *testing.T) {
	cases := map[string][]entry{
		"a path climbing out": {{header: tar.Header{Typeflag: tar.TypeReg, Name: "../escaped"}, body: "x"}},
		"an absolute path":    {{header: tar.Header{Typeflag: tar.TypeReg, Name: "/escaped"}, body: "x"}},
		"a file through a link": {
			{header: tar.Header{Typeflag: tar.TypeSymlink, Name: "out", Linkname: ".."}},
			{header: tar.Header{Typeflag: tar.TypeReg, Name: "out/escaped"}, body: "x"},
		},
		"a file over a link": {
			{header: tar.Header{Typeflag: tar.TypeSymlink, Name: "escaped-link", Linkname: "../escaped"}},
			{header: tar.Header{Typeflag: tar.TypeReg, Name: "escaped-link"}, body: "x"},
		},
		"a directory through a link": {
			{header: tar.Header{Typeflag: tar.TypeSymlink, Name: "out", Linkname: ".."}},
			{header: tar.Header{Typeflag: tar.TypeDir, Name: "out/escaped"}},
		},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			dst := filepath.Join(parent, "checkout")
			must(t, os.Mkdir(dst, 0o700))

			_, err := Unpack(context.Background(), archive(t, entries...), dst)

			assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot)
			assert.NoFileExists(t, filepath.Join(parent, "escaped"))
			assert.NoDirExists(t, filepath.Join(parent, "escaped"))
		})
	}
}

// Only files, directories and symbolic links are source: a device is refused.
func TestAnEntryThatIsNotSourceIsRefused(t *testing.T) {
	dst := t.TempDir()

	_, err := Unpack(context.Background(),
		archive(t, entry{header: tar.Header{Typeflag: tar.TypeChar, Name: "null", Devmajor: 1, Devminor: 3}}), dst)

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
	assert.NoFileExists(t, filepath.Join(dst, "null"))
}

// A stream that is cut short is an error, not a smaller source.
func TestACutStreamIsAnError(t *testing.T) {
	src := t.TempDir()
	must(t, os.WriteFile(filepath.Join(src, "a.txt"), bytes.Repeat([]byte("source "), 4096), 0o644))
	var packed bytes.Buffer
	_, err := Pack(context.Background(), src, &packed)
	must(t, err)

	_, err = Unpack(context.Background(), bytes.NewReader(packed.Bytes()[:packed.Len()/2]), t.TempDir())

	assert.Error(t, err)
}

// Packing stops when its context is done.
func TestPackingStopsWhenCanceled(t *testing.T) {
	src := t.TempDir()
	must(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Pack(ctx, src, &bytes.Buffer{})

	assert.ErrorIs(t, err, context.Canceled)
}
