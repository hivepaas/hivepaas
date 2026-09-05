package syserroruc

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// floodGuardWindow is how long one distinct error is kept out of the store after
	// it has been recorded once. A dependency going down produces the same error
	// thousands of times a minute; the first one says everything the rest would.
	floodGuardWindow = 5 * time.Minute

	// floodGuardMaxEntries bounds what the guard remembers. It is a cache of what has
	// been seen recently, not a record, so it is allowed to forget.
	floodGuardMaxEntries = 1000
)

// floodGuard keeps a burst of identical errors from becoming a burst of rows.
//
// It is deliberately in-process rather than backed by Redis. The error path is the
// one place that must keep working when dependencies do not: a Redis round trip here
// would fail exactly during the outage the guard exists for, leaving a choice between
// dropping the evidence and flooding anyway. The cost is that each replica keeps its
// own window, so N replicas record up to N rows per window instead of one - which is
// a rounding error against the thousands the guard removes.
type floodGuard struct {
	mu     sync.Mutex
	seen   map[string]time.Time
	window time.Duration
	max    int
	now    func() time.Time
}

func newFloodGuard() *floodGuard {
	return &floodGuard{
		seen:   map[string]time.Time{},
		window: floodGuardWindow,
		max:    floodGuardMaxEntries,
		now:    time.Now,
	}
}

// allow reports whether this error should be recorded, and marks it as seen when it
// is. A fingerprint not seen within the window is always allowed.
func (g *floodGuard) allow(fingerprint string) bool {
	if g == nil {
		return true
	}
	now := g.now()

	g.mu.Lock()
	defer g.mu.Unlock()

	if last, ok := g.seen[fingerprint]; ok && now.Sub(last) < g.window {
		return false
	}
	g.evictLocked(now)
	g.seen[fingerprint] = now
	return true
}

// evictLocked keeps the map bounded. Expired entries go first; if that is not enough
// the whole map is dropped, which costs one extra row per fingerprint and cannot
// degrade into unbounded growth.
func (g *floodGuard) evictLocked(now time.Time) {
	if len(g.seen) < g.max {
		return
	}
	for key, last := range g.seen {
		if now.Sub(last) >= g.window {
			delete(g.seen, key)
		}
	}
	if len(g.seen) >= g.max {
		clear(g.seen)
	}
}

// errorFingerprint identifies "the same error happening again".
//
// The stack trace is what distinguishes two failures that share a code, so it carries
// the identity; the code alone is far too coarse, since ERR_INTERNAL covers every bug
// in the process. It is hashed rather than kept whole because the guard would
// otherwise hold kilobytes per entry to compare strings it never reads.
//
// Detail and cause are deliberately left out: they carry ids and names, so including
// them would make every occurrence unique and the guard would never fire.
func errorFingerprint(errInfo *hperrors.ErrorInfo) string {
	if errInfo == nil {
		return ""
	}
	sum := sha256.Sum256([]byte(errInfo.Code + "\n" + errInfo.StackTrace))
	return hex.EncodeToString(sum[:])
}
