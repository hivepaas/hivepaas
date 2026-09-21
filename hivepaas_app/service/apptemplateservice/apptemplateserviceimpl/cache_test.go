package apptemplateserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/services/registry"
)

// The cache is addressed by content, so what the index of the revision being
// served does not name belongs to a revision that has been replaced.
func TestIndexPrunesTheFilesOfReplacedRevisions(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()

	// A file of an older revision, and one a writer is in the middle of writing.
	stale := filepath.Join(fx.cacheDir, "1111111111111111111111111111111111111111111111111111111111111111")
	partial := filepath.Join(fx.cacheDir, "2222222222222222222222222222222222222222222222222222222222222222.tmp")
	assert.NoError(t, os.WriteFile(stale, []byte("an older template"), 0o600))
	assert.NoError(t, os.WriteFile(partial, []byte("half written"), 0o600))

	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)

	assert.NoFileExists(t, stale, "a file the index does not name is of a replaced revision")
	assert.FileExists(t, partial, "a write in progress belongs to its writer")

	// Everything the index does name is kept, including the index itself.
	entry := index.FindTemplate("demo")
	_, err = fx.source.TemplateFile(ctx, entry)
	assert.NoError(t, err)
	_, err = fx.source.Icon(ctx, entry.Icon.SHA256)
	assert.NoError(t, err)

	fresh := fx.newSource()
	_, err = fresh.Index(ctx)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(fx.cacheDir, entry.File.SHA256))
	assert.FileExists(t, filepath.Join(fx.cacheDir, entry.Icon.SHA256))
	assert.FileExists(t, filepath.Join(fx.cacheDir, fx.release.Stable.Templates.IndexSHA256))
}

// Pruning is housekeeping: a directory it cannot read must not fail the request
// that happened to load the index.
func TestIndexSurvivesACacheItCannotPrune(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.source.cacheDir = func() string { return filepath.Join(fx.cacheDir, "does-not-exist") }

	index, err := fx.source.Index(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, index.FindTemplate("demo"))
}

// The pin is read from a signed envelope that is opened and verified on every
// read, and one request asks for it several times.
func TestThePinIsReadOncePerBurst(t *testing.T) {
	fx := newOfficialFixture(t)
	counting := &countingReleaseInfo{info: fx.release}
	fx.source.hpAppService = counting
	ctx := context.Background()

	_, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	_, err = fx.source.Revision(ctx)
	assert.NoError(t, err)
	_, err = fx.source.TemplateFile(ctx, mustIndex(t, fx).FindTemplate("demo"))
	assert.NoError(t, err)

	assert.Equal(t, 1, counting.calls, "the index, the revision and a file share one read of the pin")
}

func mustIndex(t *testing.T, fx *officialFixture) *templatemodel.Index {
	t.Helper()
	index, err := fx.source.Index(context.Background())
	assert.NoError(t, err)
	return index
}

func TestTemplateCacheKeepsATemplatePerFileHash(t *testing.T) {
	cache := newTemplateCache()
	first, second := &templatemodel.Template{Kind: "one"}, &templatemodel.Template{Kind: "two"}

	_, found := cache.get("aaa")
	assert.False(t, found)

	cache.put("aaa", first)
	cache.put("bbb", second)

	got, found := cache.get("aaa")
	assert.True(t, found)
	assert.Same(t, first, got)
	got, found = cache.get("bbb")
	assert.True(t, found)
	assert.Same(t, second, got)
}

// A template file that changes gets a new hash, so the old entry is never
// served in its place - which is what makes the cache need no invalidation.
func TestTemplateCacheIsBoundedAndDropsWhatWasUsedLongestAgo(t *testing.T) {
	cache := newTemplateCache()
	for i := range templateCacheMax {
		cache.put(key(i), &templatemodel.Template{Kind: key(i)})
		cache.entries[key(i)].usedAt = time.Now().Add(time.Duration(i) * time.Second)
	}
	assert.Len(t, cache.entries, templateCacheMax)

	cache.put("newest", &templatemodel.Template{Kind: "newest"})

	assert.Len(t, cache.entries, templateCacheMax)
	_, found := cache.get(key(0))
	assert.False(t, found, "the one used longest ago made room")
	_, found = cache.get("newest")
	assert.True(t, found)
}

func TestTemplateCacheIsSafeWhenThereIsNone(t *testing.T) {
	var cache *templateCache
	_, found := cache.get("aaa")
	assert.False(t, found)
	assert.NotPanics(t, func() { cache.put("aaa", &templatemodel.Template{}) })
}

func key(i int) string {
	return string(rune('a'+i%26)) + string(rune('a'+i/26))
}

// Without this the map holds one entry per repository anybody ever opened, each
// carrying thousands of tag strings.
func TestImageTagsCacheDropsExpiredEntriesWhenItStoresOne(t *testing.T) {
	cache := newImageTagsCache()
	cache.put("docker.io/library/postgres", &registry.ListTagsResult{Tags: []string{"18"}})
	cache.entries["docker.io/library/postgres"].readAt = time.Now().Add(-imageTagsCacheTTL - time.Minute)

	cache.put("docker.io/library/mariadb", &registry.ListTagsResult{Tags: []string{"11.8"}})

	assert.Len(t, cache.entries, 1)
	_, found := cache.get("docker.io/library/mariadb")
	assert.True(t, found)
}

type countingReleaseInfo struct {
	fakeReleaseInfo
	calls int
}

func (c *countingReleaseInfo) GetAppReleaseInfo(ctx context.Context) (*hpappservice.AppReleaseInfo, error) {
	c.calls++
	return c.fakeReleaseInfo.GetAppReleaseInfo(ctx)
}
