package artifactagentuc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
)

var content = []byte("#!/bin/sh\necho wal-g\n")

func hashOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// fileServer serves body at every path, and counts what it is asked.
type fileServer struct {
	*httptest.Server
	hits atomic.Int32
}

func serve(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *fileServer {
	t.Helper()
	server := &fileServer{}
	server.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.hits.Add(1)
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func serveBody(t *testing.T, body []byte) *fileServer {
	t.Helper()
	return serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	})
}

// newTestUC keeps its cache in a directory of the test's, which the host sees as
// /var/lib/hivepaas, and trusts the test servers' certificate.
func newTestUC(t *testing.T, servers ...*fileServer) (*UC, string) {
	t.Helper()
	cacheDir := filepath.Join(t.TempDir(), "cache")
	logger, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)
	uc := newUC(logger, cacheDir, "/var/lib/hivepaas/cache/artifacts/sha256")
	if len(servers) > 0 {
		uc.client = servers[0].Client()
	}
	return uc, cacheDir
}

func ensureReq(url string) *artifactagentdto.EnsureReq {
	return &artifactagentdto.EnsureReq{URL: url, SHA256: hashOf(content), Size: int64(len(content))}
}

// filesIn lists what a directory holds, temp files included.
func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	assert.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestEnsureDownloadsAndKeepsAFileByItsHash(t *testing.T) {
	source := serveBody(t, content)
	uc, cacheDir := newTestUC(t, source)

	resp, err := uc.Ensure(context.Background(), ensureReq(source.URL+"/wal-g"))

	assert.NoError(t, err)
	assert.True(t, resp.Downloaded)
	assert.Equal(t, "/var/lib/hivepaas/cache/artifacts/sha256/"+hashOf(content), resp.HostPath)
	data, err := os.ReadFile(filepath.Join(cacheDir, hashOf(content)))
	assert.NoError(t, err)
	assert.Equal(t, content, data)
	info, err := os.Stat(filepath.Join(cacheDir, hashOf(content)))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm(), "a binary a non-root container runs")
}

func TestEnsureDoesNotDownloadAFileItHas(t *testing.T) {
	source := serveBody(t, content)
	uc, _ := newTestUC(t, source)
	_, err := uc.Ensure(context.Background(), ensureReq(source.URL))
	assert.NoError(t, err)

	resp, err := uc.Ensure(context.Background(), ensureReq(source.URL))

	assert.NoError(t, err)
	assert.False(t, resp.Downloaded)
	assert.Equal(t, int32(1), source.hits.Load())
}

// A file that is there but is not what its name says - changed on the host, or
// cut short by a disk that filled - is checked once, found wrong, and replaced.
func TestEnsureReplacesAFileThatIsNotWhatItsHashSays(t *testing.T) {
	source := serveBody(t, content)
	uc, cacheDir := newTestUC(t, source)
	assert.NoError(t, os.MkdirAll(cacheDir, 0o755))
	tampered := []byte("#!/bin/sh\nrm -rf /tmp/x\n")[:len(content)] // the same size, other bytes
	assert.NoError(t, os.WriteFile(filepath.Join(cacheDir, hashOf(content)), tampered, 0o755))

	resp, err := uc.Ensure(context.Background(), ensureReq(source.URL))

	assert.NoError(t, err)
	assert.True(t, resp.Downloaded)
	data, _ := os.ReadFile(filepath.Join(cacheDir, hashOf(content)))
	assert.Equal(t, content, data)
}

func TestEnsureRefusesAFileWithAnotherHashAndKeepsNothing(t *testing.T) {
	source := serveBody(t, []byte("#!/bin/sh\necho evil!\n"))
	uc, cacheDir := newTestUC(t, source)

	_, err := uc.Ensure(context.Background(), ensureReq(source.URL))

	assert.ErrorIs(t, err, hperrors.ErrArtifactHashMismatch)
	assert.Empty(t, filesIn(t, cacheDir), "no file, and no temp file")
}

func TestEnsureKeepsNothingOfADownloadCutShort(t *testing.T) {
	source := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		_, _ = w.Write(content[:5])
		w.(http.Flusher).Flush()
		conn, _, _ := w.(http.Hijacker).Hijack()
		_ = conn.Close()
	})
	uc, cacheDir := newTestUC(t, source)

	_, err := uc.Ensure(context.Background(), ensureReq(source.URL))

	assert.ErrorIs(t, err, hperrors.ErrArtifactDownloadFailed)
	assert.Empty(t, filesIn(t, cacheDir))
}

func TestEnsureStopsReadingPastTheSize(t *testing.T) {
	larger := append(append([]byte{}, content...), make([]byte, 1<<20)...)
	for name, server := range map[string]func(t *testing.T) *fileServer{
		"saying so": func(t *testing.T) *fileServer { return serveBody(t, larger) },
		"not saying so": func(t *testing.T) *fileServer {
			return serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(larger) })
		},
	} {
		t.Run(name, func(t *testing.T) {
			source := server(t)
			uc, cacheDir := newTestUC(t, source)

			_, err := uc.Ensure(context.Background(), ensureReq(source.URL))

			assert.ErrorIs(t, err, hperrors.ErrArtifactDownloadFailed)
			assert.Empty(t, filesIn(t, cacheDir))
		})
	}
}

func TestEnsureDownloadsOnceForCallsAtOnce(t *testing.T) {
	release := make(chan struct{})
	source := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		<-release
		_, _ = w.Write(content)
	})
	uc, _ := newTestUC(t, source)

	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = uc.Ensure(context.Background(), ensureReq(source.URL))
		}()
	}
	close(release)
	wg.Wait()

	for _, err := range errs {
		assert.NoError(t, err)
	}
	assert.Equal(t, int32(1), source.hits.Load())
}

func TestEnsureTriesTheMirrorFirst(t *testing.T) {
	source := serveBody(t, content)
	mirror := serveBody(t, content)
	uc, _ := newTestUC(t, source)
	req := ensureReq(source.URL)
	req.MirrorURL = mirror.URL + "/" + hashOf(content)

	_, err := uc.Ensure(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), mirror.hits.Load())
	assert.Zero(t, source.hits.Load())
}

// A mirror that does not have the file, or has another one, is passed over for
// the source: the hash decides either way.
func TestEnsureFallsBackToTheSource(t *testing.T) {
	for name, mirrorBody := range map[string]func(w http.ResponseWriter, r *http.Request){
		"missing": func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) },
		"wrong":   func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("something else")) },
	} {
		t.Run(name, func(t *testing.T) {
			source := serveBody(t, content)
			mirror := serve(t, mirrorBody)
			uc, _ := newTestUC(t, source)
			req := ensureReq(source.URL)
			req.MirrorURL = mirror.URL + "/" + hashOf(content)

			_, err := uc.Ensure(context.Background(), req)

			assert.NoError(t, err)
			assert.Equal(t, int32(1), source.hits.Load())
		})
	}
}

func TestEnsureRefusesWhatItCannotTrust(t *testing.T) {
	uc, _ := newTestUC(t)
	for name, req := range map[string]*artifactagentdto.EnsureReq{
		"an http url":    {URL: "http://example.com/wal-g", SHA256: hashOf(content), Size: 1},
		"an http mirror": {URL: "https://example.com/wal-g", MirrorURL: "http://m/x", SHA256: hashOf(content), Size: 1},
		"a short hash":   {URL: "https://example.com/wal-g", SHA256: "abc", Size: 1},
		"a path in hash": {URL: "https://example.com/wal-g", SHA256: "../../etc/passwd", Size: 1},
		"no size":        {URL: "https://example.com/wal-g", SHA256: hashOf(content)},
		"too big":        {URL: "https://example.com/wal-g", SHA256: hashOf(content), Size: maxSize + 1},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Ensure(context.Background(), req)
			assert.ErrorIs(t, err, hperrors.ErrArtifactInvalid)
		})
	}
}
