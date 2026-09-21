package apptemplateserviceimpl

import (
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/services/registry"
)

// imageTagsCacheTTL is how long a repository's tag list is reused. Tags change
// when an image is published - hours or days apart - and the cost of being a few
// minutes behind is one stale row in a list somebody is reading anyway.
const imageTagsCacheTTL = 10 * time.Minute

type imageTagsCacheEntry struct {
	result *registry.ListTagsResult
	readAt time.Time
}

// imageTagsCache keeps one tag list per repository. Only a successful read is
// stored: a rate-limited or unreachable registry must not be remembered as an
// empty repository.
type imageTagsCache struct {
	mu      sync.Mutex
	entries map[string]*imageTagsCacheEntry
}

func newImageTagsCache() *imageTagsCache {
	return &imageTagsCache{entries: map[string]*imageTagsCacheEntry{}}
}

func (c *imageTagsCache) get(key string) (*registry.ListTagsResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.entries[key]
	if !found || time.Since(entry.readAt) > imageTagsCacheTTL {
		return nil, false
	}
	return entry.result, true
}

func (c *imageTagsCache) put(key string, result *registry.ListTagsResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Drop what has expired while here. Without this the map only ever grows: one
	// entry for every repository anybody has ever opened the tag list of, each
	// holding thousands of tag strings long after it stopped being read.
	now := time.Now()
	for other, entry := range c.entries {
		if now.Sub(entry.readAt) > imageTagsCacheTTL {
			delete(c.entries, other)
		}
	}
	c.entries[key] = &imageTagsCacheEntry{result: result, readAt: now}
}
