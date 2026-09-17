package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// fakeRegistry answers the way a V2 registry does: a challenge first, then tags in
// pages, with a Link header pointing at the next one.
type fakeRegistry struct {
	mu sync.Mutex

	server    *httptest.Server
	pages     [][]string
	status    int
	tokenSeen []string
	tokens    int
}

func newFakeRegistry(t *testing.T, pages ...[]string) *fakeRegistry {
	t.Helper()
	fake := &fakeRegistry{pages: pages}
	fake.server = httptest.NewServer(fake)
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.URL.Path == "/token" {
		f.tokens++
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"token":"test-token"}`)
		return
	}
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	if r.Header.Get("Authorization") == "" {
		w.Header().Set("WWW-Authenticate",
			fmt.Sprintf(`Bearer realm="%s/token",service="fake",scope="repository:library/postgres:pull"`,
				f.server.URL))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	f.tokenSeen = append(f.tokenSeen, r.Header.Get("Authorization"))

	page := 0
	if last := r.URL.Query().Get("last"); last != "" {
		_, _ = fmt.Sscanf(last, "page-%d", &page)
	}
	if page >= len(f.pages) {
		_, _ = fmt.Fprint(w, `{"name":"library/postgres","tags":[]}`)
		return
	}
	if page+1 < len(f.pages) {
		w.Header().Set("Link",
			fmt.Sprintf(`</v2/library/postgres/tags/list?n=100&last=page-%d>; rel="next"`, page+1))
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"name":"library/postgres","tags":["%s"]}`, strings.Join(f.pages[page], `","`))
}

func (f *fakeRegistry) client() *Client {
	client := New()
	client.endpoint = func(string) string { return f.server.URL }
	return client
}

func testRef() Reference {
	return Reference{Host: "registry-1.docker.io", Name: "library/postgres"}
}

func TestListTagsFollowsTheTokenChallengeAndPages(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6", "18.7"}, []string{"19.0"})

	result, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.NoError(t, err)
	assert.Equal(t, []string{"18.6", "18.7", "19.0"}, result.Tags)
	assert.False(t, result.Truncated)
	assert.Equal(t, 1, fake.tokens, "the token is fetched once and reused for the next page")
	assert.Equal(t, []string{"Bearer test-token", "Bearer test-token"}, fake.tokenSeen)
}

func TestListTagsStopsAtTheCap(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6", "18.7"}, []string{"19.0", "19.1"})

	result, err := fake.client().ListTags(context.Background(), testRef(), 3)

	assert.NoError(t, err)
	assert.Len(t, result.Tags, 3)
	assert.True(t, result.Truncated, "the caller has to know the answer is partial")
}

func TestListTagsReportsRateLimiting(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	fake.status = http.StatusTooManyRequests

	_, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrRegistryRateLimited)
}

func TestListTagsReportsAMissingRepository(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	fake.status = http.StatusNotFound

	_, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestListTagsReportsAnUnreachableRegistry(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	client := fake.client()
	fake.server.Close()

	_, err := client.ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrRegistryUnavailable)
}
