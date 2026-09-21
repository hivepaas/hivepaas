package apptemplateserviceimpl

import (
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// templateCacheMax is how many decoded templates are kept. A template is a few
// kilobytes of YAML and perhaps twice that decoded, so this is a small amount of
// memory for the templates anyone is actually opening; the rest are decoded
// again when they are asked for.
const templateCacheMax = 64

// templateCache keeps templates that have already been decoded, by the sha256 of
// the file they were decoded from.
//
// The key is what makes it safe: a template file that changes has a different
// hash, so a revision moving on is a new key rather than a stale value, and
// nothing has to be invalidated when the pin changes. There is no TTL for the
// same reason.
//
// What it saves is real on the deploy path: creating an app from a template with
// three dependencies reads and decodes four files, and the same four on every
// request that renders it.
//
// A cached template is shared, so callers must read it and not modify it -
// templaterender copies the app tree before it substitutes anything.
type templateCache struct {
	mu      sync.Mutex
	entries map[string]*templateCacheEntry
}

type templateCacheEntry struct {
	tmpl   *templatemodel.Template
	usedAt time.Time
}

func newTemplateCache() *templateCache {
	return &templateCache{entries: map[string]*templateCacheEntry{}}
}

// get and put are safe on a nil cache: it is an accelerator, not a dependency,
// so a service built without one simply decodes every time.
func (c *templateCache) get(sha256Hex string) (*templatemodel.Template, bool) {
	if c == nil || sha256Hex == "" {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.entries[sha256Hex]
	if !found {
		return nil, false
	}
	entry.usedAt = time.Now()
	return entry.tmpl, true
}

// put stores a decoded template, dropping the one used longest ago when the
// cache is full.
func (c *templateCache) put(sha256Hex string, tmpl *templatemodel.Template) {
	if c == nil || sha256Hex == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.entries[sha256Hex]; !found && len(c.entries) >= templateCacheMax {
		c.evictOldest()
	}
	c.entries[sha256Hex] = &templateCacheEntry{tmpl: tmpl, usedAt: time.Now()}
}

// evictOldest is called with the lock held.
func (c *templateCache) evictOldest() {
	var oldestKey string
	var oldestAt time.Time
	for key, entry := range c.entries {
		if oldestKey == "" || entry.usedAt.Before(oldestAt) {
			oldestKey, oldestAt = key, entry.usedAt
		}
	}
	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}
