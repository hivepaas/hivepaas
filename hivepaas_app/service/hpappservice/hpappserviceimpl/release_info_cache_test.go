package hpappserviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testReleaseURL      = "https://example.test/release/release.signed.json"
	testReleaseURLOther = "https://example.test/main/release.signed.json"
)

var errTestFetch = errors.New("fetch failed")

// fakeReleaseSource serves a fixed envelope and counts the fetches.
type fakeReleaseSource struct {
	envelope []byte
	err      error
	fetches  int
}

func (f *fakeReleaseSource) fetch(_ context.Context, _ string) ([]byte, error) {
	f.fetches++
	return f.envelope, f.err
}

func acceptAll([]byte) error { return nil }

func TestReleaseInfoCache(t *testing.T) {
	ctx := context.Background()
	t0 := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)

	t.Run("reuses the envelope within the TTL", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("v1")}

		for _, at := range []time.Time{t0, t0.Add(time.Minute), t0.Add(releaseInfoCacheTTL - time.Second)} {
			got, err := cache.get(ctx, testReleaseURL, at, src.fetch, acceptAll)
			assert.NoError(t, err)
			assert.Equal(t, []byte("v1"), got)
		}
		assert.Equal(t, 1, src.fetches)
	})

	t.Run("fetches again once the TTL has passed", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("v1")}
		_, _ = cache.get(ctx, testReleaseURL, t0, src.fetch, acceptAll)

		src.envelope = []byte("v2")
		got, err := cache.get(ctx, testReleaseURL, t0.Add(releaseInfoCacheTTL), src.fetch, acceptAll)
		assert.NoError(t, err)
		assert.Equal(t, []byte("v2"), got)
		assert.Equal(t, 2, src.fetches)
	})

	t.Run("an entry for one URL does not answer for another", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("release")}
		_, _ = cache.get(ctx, testReleaseURL, t0, src.fetch, acceptAll)

		src.envelope = []byte("main")
		got, err := cache.get(ctx, testReleaseURLOther, t0.Add(time.Minute), src.fetch, acceptAll)
		assert.NoError(t, err)
		assert.Equal(t, []byte("main"), got)
		assert.Equal(t, 2, src.fetches)
	})

	t.Run("a failed fetch is not remembered", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{err: errTestFetch}
		_, err := cache.get(ctx, testReleaseURL, t0, src.fetch, acceptAll)
		assert.ErrorIs(t, err, errTestFetch)

		src.err, src.envelope = nil, []byte("v1")
		got, err := cache.get(ctx, testReleaseURL, t0.Add(time.Second), src.fetch, acceptAll)
		assert.NoError(t, err)
		assert.Equal(t, []byte("v1"), got)
		assert.Equal(t, 2, src.fetches)
	})

	t.Run("an envelope that is not accepted is not kept", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("forged")}
		reject := func([]byte) error { return errTestFetch }

		got, err := cache.get(ctx, testReleaseURL, t0, src.fetch, reject)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errTestFetch)

		src.envelope = []byte("genuine")
		got, err = cache.get(ctx, testReleaseURL, t0.Add(time.Second), src.fetch, acceptAll)
		assert.NoError(t, err)
		assert.Equal(t, []byte("genuine"), got, "the rejected envelope must not be served from the cache")
		assert.Equal(t, 2, src.fetches)
	})

	t.Run("a failure after expiry does not serve the stale entry", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("v1")}
		_, _ = cache.get(ctx, testReleaseURL, t0, src.fetch, acceptAll)

		src.err = errTestFetch
		got, err := cache.get(ctx, testReleaseURL, t0.Add(releaseInfoCacheTTL), src.fetch, acceptAll)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errTestFetch)
	})

	t.Run("a clock that went backwards fetches again", func(t *testing.T) {
		var cache releaseInfoCache
		src := &fakeReleaseSource{envelope: []byte("v1")}
		_, _ = cache.get(ctx, testReleaseURL, t0, src.fetch, acceptAll)
		_, _ = cache.get(ctx, testReleaseURL, t0.Add(-time.Hour), src.fetch, acceptAll)
		assert.Equal(t, 2, src.fetches)
	})
}
