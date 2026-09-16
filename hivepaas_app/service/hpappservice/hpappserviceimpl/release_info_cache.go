package hpappserviceimpl

import (
	"context"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// releaseInfoCacheTTL is how long a fetched release.signed.json is reused before
// it is fetched again.
//
// Release info is read on every visit to the update page and again when an
// update is started; without a cache each is a request to GitHub, which rate
// limits by IP and is shared by every installation behind one. A release takes
// at most this long, plus GitHub's own few minutes of CDN caching, to be seen.
const releaseInfoCacheTTL = 30 * time.Minute

// releaseInfoFetchTimeout bounds one fetch of release.signed.json. The cache holds
// its lock across the fetch, so without a bound a request GitHub never answers
// would hold every other reader of release info with it.
const releaseInfoFetchTimeout = 10 * time.Second

// releaseInfoCache keeps the last release.signed.json that verified, per URL.
//
// It keeps the envelope, not the decoded info: every read opens it again, so the
// signatures are checked on each use exactly as on a fetch, and each caller gets
// a fresh AppReleaseInfo it is free to modify. Opening is cheap next to a request.
//
// Only an envelope that verified is kept. A failed fetch or a bad envelope is not
// remembered, so the next read tries again rather than serving the failure for
// the whole TTL.
//
// The zero value is ready to use.
type releaseInfoCache struct {
	// mu is held across the fetch as well, so concurrent readers of a stale
	// entry wait for one request rather than each making their own.
	mu        sync.Mutex
	url       string
	envelope  []byte
	fetchedAt time.Time
}

// get returns the cached envelope for url if it is younger than the TTL, and
// otherwise fetches it and passes it to accept, keeping it only if accept
// returns no error.
func (c *releaseInfoCache) get(
	ctx context.Context,
	url string,
	now time.Time,
	fetch func(ctx context.Context, url string) ([]byte, error),
	accept func(envelope []byte) error,
) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// The URL is part of the key: it follows the environment, and an entry read
	// from one branch must not answer for the other.
	if c.envelope != nil && c.url == url && now.Sub(c.fetchedAt) < releaseInfoCacheTTL && !now.Before(c.fetchedAt) {
		return c.envelope, nil
	}

	envelope, err := fetch(ctx, url)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = accept(envelope); err != nil {
		return nil, hperrors.Wrap(err)
	}

	c.url = url
	c.envelope = envelope
	c.fetchedAt = now
	return envelope, nil
}
