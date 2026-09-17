package apptemplateserviceimpl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	testRepo       = "hivepaas/app-templates"
	testCommit     = "0123456789abcdef0123456789abcdef01234567"
	testBetaCommit = "89abcdef0123456789abcdef0123456789abcdef"
	testRepoDir    = "testdata/repo"
)

type fakeReleaseInfo struct {
	hpappservice.Service
	info *hpappservice.AppReleaseInfo
}

func (f *fakeReleaseInfo) GetAppReleaseInfo(context.Context) (*hpappservice.AppReleaseInfo, error) {
	return f.info, nil
}

// templatesServer stands in for raw.githubusercontent.com.
type templatesServer struct {
	mu     sync.Mutex
	files  map[string][]byte
	hits   map[string]int
	server *httptest.Server
}

func (ts *templatesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.hits[r.URL.Path]++
	data, ok := ts.files[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}

func (ts *templatesServer) set(path string, data []byte) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.files["/"+testRepo+"/"+testCommit+"/"+path] = data
}

func (ts *templatesServer) hitCount(path string) int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.hits["/"+testRepo+"/"+testCommit+"/"+path]
}

type officialFixture struct {
	source    *officialSource
	server    *templatesServer
	cacheDir  string
	indexData []byte
	template  []byte
	icon      []byte
	release   *hpappservice.AppReleaseInfo
}

func useEnv(t *testing.T, env string) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: env})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func pinOf(commit string, indexData []byte) *hpappservice.ReleaseInfo {
	sum := sha256.Sum256(indexData)
	return &hpappservice.ReleaseInfo{ReleaseInfo: base.ReleaseInfo{
		AppVersion: "v0.1.1",
		Templates:  &base.TemplatesRef{Repo: testRepo, Commit: commit, IndexSHA256: hex.EncodeToString(sum[:])},
	}}
}

func newOfficialFixture(t *testing.T) *officialFixture {
	t.Helper()
	useEnv(t, config.EnvProd)

	repo, problems, err := templaterepo.Load(os.DirFS(testRepoDir))
	assert.NoError(t, err)
	assert.Empty(t, problems)
	index, err := templaterepo.BuildIndex(repo)
	assert.NoError(t, err)
	indexData, err := templaterepo.MarshalIndex(index)
	assert.NoError(t, err)

	fx := &officialFixture{
		server:    &templatesServer{files: map[string][]byte{}, hits: map[string]int{}},
		cacheDir:  t.TempDir(),
		indexData: indexData,
		template:  repo.FindTemplate("demo").Content,
		icon:      repo.Icons["icons/demo.svg"].Content,
		release:   &hpappservice.AppReleaseInfo{Stable: pinOf(testCommit, indexData)},
	}
	fx.server.server = httptest.NewServer(fx.server)
	t.Cleanup(fx.server.server.Close)
	fx.server.set(templaterepo.IndexFile, indexData)
	fx.server.set("templates/demo.yaml", fx.template)
	fx.server.set("icons/demo.svg", fx.icon)
	fx.source = fx.newSource()
	return fx
}

// newSource is a fresh source - no in-memory index - over the same server and
// disk cache, the way a restarted process sees them.
func (fx *officialFixture) newSource() *officialSource {
	source := newOfficialSource(&fakeReleaseInfo{info: fx.release})
	source.baseURL = fx.server.server.URL
	source.cacheDir = func() string { return fx.cacheDir }
	return source
}

func TestOfficialSourceServesVerifiedFiles(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()

	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")

	template, err := fx.source.TemplateFile(ctx, entry)
	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)

	icon, err := fx.source.Icon(ctx, entry.Icon.SHA256)
	assert.NoError(t, err)
	assert.Equal(t, fx.icon, icon)

	revision, err := fx.source.Revision(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testCommit, revision)

	cached, err := os.ReadFile(filepath.Join(fx.cacheDir, entry.File.SHA256))
	assert.NoError(t, err)
	assert.Equal(t, fx.template, cached, "a verified file is cached under its hash")
}

func TestOfficialSourceServesTheCacheWhenGitHubIsDown(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()
	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	_, err = fx.source.TemplateFile(ctx, index.FindTemplate("demo"))
	assert.NoError(t, err)

	fx.server.server.Close()
	restarted := fx.newSource()

	index, err = restarted.Index(ctx)
	assert.NoError(t, err)
	template, err := restarted.TemplateFile(ctx, index.FindTemplate("demo"))
	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)
}

func TestOfficialSourceRefusesAFileThatDoesNotMatchTheIndex(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.server.set("templates/demo.yaml", append(bytes.Clone(fx.template), []byte("# tampered\n")...))
	ctx := context.Background()

	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")
	_, err = fx.source.TemplateFile(ctx, entry)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVerificationFailed)
	assert.NoFileExists(t, filepath.Join(fx.cacheDir, entry.File.SHA256), "nothing unverified is cached")
}

func TestOfficialSourceRefusesAnIndexThatDoesNotMatchThePin(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.server.set(templaterepo.IndexFile, bytes.Replace(fx.indexData, []byte("Demo"), []byte("Evil"), 1))

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVerificationFailed)
}

func TestOfficialSourceRefetchesACorruptedCacheEntry(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()
	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")
	_, err = fx.source.TemplateFile(ctx, entry)
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(fx.cacheDir, entry.File.SHA256), []byte("corrupt"), 0o600))
	template, err := fx.source.TemplateFile(ctx, entry)

	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)
	assert.Equal(t, 2, fx.server.hitCount("templates/demo.yaml"))
}

func TestOfficialSourceRefusesAnOversizedFile(t *testing.T) {
	fx := newOfficialFixture(t)
	huge := []byte(strings.Repeat("x", templaterepo.MaxIndexSize+1))
	fx.release.Stable = pinOf(testCommit, huge)
	fx.server.set(templaterepo.IndexFile, huge)

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateFileTooLarge)
}

func TestOfficialSourceWithoutAPin(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.release.Stable.Templates = nil

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplatesUnavailable)
}

func TestOfficialSourceReadsTheBetaPinInTheBetaEnv(t *testing.T) {
	fx := newOfficialFixture(t)
	useEnv(t, config.EnvBeta)
	fx.release.Beta = pinOf(testBetaCommit, fx.indexData)

	revision, err := fx.source.Revision(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, testBetaCommit, revision)
}

func TestOfficialSourceServesOnlyIconsTheIndexLists(t *testing.T) {
	fx := newOfficialFixture(t)
	index, err := fx.source.Index(context.Background())
	assert.NoError(t, err)

	_, err = fx.source.Icon(context.Background(), index.FindTemplate("demo").File.SHA256)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}
