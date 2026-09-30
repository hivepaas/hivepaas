package repocheckoutserviceimpl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filearchiver"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
)

// fakeFiles keeps files in memory, by path.
type fakeFiles struct {
	fileservice.Service
	files   map[string][]byte
	volumes map[string]*entity.Setting
}

func (f *fakeFiles) Open(_ context.Context, _ database.IDB, file *entity.File) (io.ReadCloser, error) {
	content, ok := f.files[file.Path]
	if !ok {
		return nil, hperrors.NewNotFound("File")
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

type memWriter struct {
	bytes.Buffer
	keep func([]byte)
}

func (w *memWriter) Close() error { w.keep(w.Bytes()); return nil }
func (w *memWriter) Abort(error)  {}

func (f *fakeFiles) Create(_ context.Context, _ database.IDB, file *entity.File) (fileservice.FileWriter, error) {
	return &memWriter{keep: func(b []byte) { f.files[file.Path] = append([]byte(nil), b...) }}, nil
}

func (f *fakeFiles) ProjectVolume(_ context.Context, _ database.IDB, projectID string) (*entity.Setting, error) {
	if v := f.volumes[projectID]; v != nil {
		return v, nil
	}
	return nil, hperrors.NewNotFound("Project default volume")
}

// The archive a cache is made into is stored through the file layer, and read
// back the same way.
func TestARepositoryCacheArchiveGoesThroughTheFileLayer(t *testing.T) {
	files := &fakeFiles{files: map[string][]byte{}}
	s := &service{fileService: files}
	archive := filepath.Join(t.TempDir(), "made.tar.lz4")
	assert.NoError(t, os.WriteFile(archive, []byte("archive bytes"), 0o600))
	file := &entity.File{Path: ".hivepaas/cache/repos/01J.tar.lz4"}

	size, err := s.storeCacheArchive(context.Background(), archive, file)
	assert.NoError(t, err)
	assert.Equal(t, int64(len("archive bytes")), size)
	assert.Equal(t, "archive bytes", string(files.files[".hivepaas/cache/repos/01J.tar.lz4"]))

	fetched, err := s.fetchCacheArchive(context.Background(), nil, file, t.TempDir())
	assert.NoError(t, err)
	content, _ := os.ReadFile(fetched)
	assert.Equal(t, "archive bytes", string(content))
}

// The cache of a project goes to its default volume, in HivePaaS's directory.
func TestARepositoryCacheIsPlacedInTheProjectVolume(t *testing.T) {
	files := &fakeFiles{volumes: map[string]*entity.Setting{"p1": {ID: "vol-p1"}}}
	s := &service{fileService: files}
	file := &entity.File{Name: "01J.abcd.tar.lz4"}

	err := s.placeRepoCache(context.Background(), "p1", file)

	assert.NoError(t, err)
	assert.Equal(t, "vol-p1", file.StorageID)
	assert.Equal(t, ".hivepaas/cache/repos/01J.abcd.tar.lz4", file.Path)
}

// A project without a default volume keeps no cache; the checkout goes on.
func TestARepositoryCacheNeedsAProjectVolume(t *testing.T) {
	s := &service{fileService: &fakeFiles{}}

	err := s.placeRepoCache(context.Background(), "p-none", &entity.File{Name: "x"})

	assert.True(t, errors.Is(err, hperrors.ErrNotFound))
}

// The copy a cache is unpacked from keeps the archive's whole extension. The
// archiver tells the format from the name: with only the last part (".lz4" of
// ".tar.lz4") it took the cache for a single compressed file, unpacked no
// repository, and every deployment fell back to a fresh clone.
func TestAFetchedCacheArchiveKeepsItsFormat(t *testing.T) {
	name := "01JAB9XED0GTXBSQDFVYAJ8WF1.9f3c1de0" + repoCacheArchiveFormat.FileExtDefault()
	file := &entity.File{Name: name, Path: ".hivepaas/cache/repos/" + name}
	files := &fakeFiles{files: map[string][]byte{file.Path: []byte("archive bytes")}}
	s := &service{fileService: files}

	fetched, err := s.fetchCacheArchive(context.Background(), nil, file, t.TempDir())

	assert.NoError(t, err)
	assert.Equal(t, repoCacheArchiveFormat, filearchiver.DetectArchiveFormat(fetched))
}

// A checkout stored as a cache comes back a git repository: packed, stored
// through the file layer, fetched and unpacked with the real archiver.
func TestACachedCheckoutComesBackAGitRepository(t *testing.T) {
	for _, tool := range []string{"tar", "lz4"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	checkout := filepath.Join(t.TempDir(), "checkout")
	assert.NoError(t, os.MkdirAll(filepath.Join(checkout, ".git"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(checkout, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(checkout, "main.go"), []byte("package main\n"), 0o600))

	name := "01JAB9XED0GTXBSQDFVYAJ8WF1.9f3c1de0" + repoCacheArchiveFormat.FileExtDefault()
	file := &entity.File{Name: name, Path: ".hivepaas/cache/repos/" + name}
	s := &service{fileService: &fakeFiles{files: map[string][]byte{}}}

	packed := filepath.Join(t.TempDir(), name)
	_, err := filearchiver.Compress(checkout, packed, repoCacheArchiveFormat, repoCacheArchiveCompressionLevel)
	if !assert.NoError(t, err) {
		return
	}
	_, err = s.storeCacheArchive(context.Background(), packed, file)
	assert.NoError(t, err)

	fetched, err := s.fetchCacheArchive(context.Background(), nil, file, t.TempDir())
	if !assert.NoError(t, err) {
		return
	}
	restored := filepath.Join(t.TempDir(), "checkout")
	_, err = filearchiver.Decompress(fetched, restored, filearchiver.ArchiveFormatAuto)
	assert.NoError(t, err)

	head, err := os.ReadFile(filepath.Join(restored, ".git", "HEAD"))
	assert.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/main\n", string(head))
}
