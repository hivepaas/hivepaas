package syserroruc

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func testGuard(now *time.Time) *floodGuard {
	g := newFloodGuard()
	g.now = func() time.Time { return *now }
	return g
}

func TestFloodGuard_CollapsesRepeatsWithinTheWindow(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	g := testGuard(&now)

	assert.True(t, g.allow("fp"), "the first occurrence must be recorded")
	for range 1000 {
		assert.False(t, g.allow("fp"), "a repeat inside the window must be dropped")
	}

	// Just before the window closes it is still a repeat.
	now = now.Add(floodGuardWindow - time.Nanosecond)
	assert.False(t, g.allow("fp"))

	// Once the window has passed the error is worth recording again.
	now = now.Add(time.Nanosecond)
	assert.True(t, g.allow("fp"))
}

// Collapsing must apply per error, or one noisy failure would hide every other.
func TestFloodGuard_DistinctErrorsAreIndependent(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	g := testGuard(&now)

	assert.True(t, g.allow("a"))
	assert.True(t, g.allow("b"))
	assert.False(t, g.allow("a"))
	assert.False(t, g.allow("b"))
}

// The guard is a cache, so it must stay bounded no matter how many distinct errors
// arrive - otherwise it becomes the leak it was added to prevent.
func TestFloodGuard_StaysBounded(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	g := testGuard(&now)

	for i := range floodGuardMaxEntries * 3 {
		g.allow(fmt.Sprintf("fp-%d", i))
	}
	assert.LessOrEqual(t, len(g.seen), floodGuardMaxEntries)
}

// Entries that have expired should be reclaimed without throwing away live ones.
func TestFloodGuard_EvictsExpiredBeforeClearing(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	g := testGuard(&now)

	for i := range floodGuardMaxEntries - 1 {
		g.allow(fmt.Sprintf("old-%d", i))
	}
	now = now.Add(floodGuardWindow)

	// This one lands after every earlier entry has expired, so the eviction has room
	// to work and does not need to clear the map.
	assert.True(t, g.allow("fresh"))
	assert.True(t, g.allow("fresh2"))
	assert.Less(t, len(g.seen), floodGuardMaxEntries)
	assert.Contains(t, g.seen, "fresh")
}

// A nil guard must not stop errors being recorded: a UC built without one should
// degrade to "record everything", never to "record nothing".
func TestFloodGuard_NilAllowsEverything(t *testing.T) {
	var g *floodGuard
	assert.True(t, g.allow("fp"))
	assert.True(t, g.allow("fp"))
}

func TestFloodGuard_ConcurrentUseIsSafe(t *testing.T) {
	now := time.Now()
	g := testGuard(&now)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if g.allow("same") {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Exactly one caller may pass, or the lock is not doing its job.
	assert.Equal(t, 1, allowed)
}

func TestErrorFingerprint(t *testing.T) {
	base := &hperrors.ErrorInfo{Code: "ERR_INTERNAL", StackTrace: "main.go:1"}

	assert.Equal(t, errorFingerprint(base), errorFingerprint(
		&hperrors.ErrorInfo{Code: "ERR_INTERNAL", StackTrace: "main.go:1"}))

	// A different stack is a different failure even under the same code.
	assert.NotEqual(t, errorFingerprint(base), errorFingerprint(
		&hperrors.ErrorInfo{Code: "ERR_INTERNAL", StackTrace: "other.go:9"}))

	// A different code is a different failure even from the same place.
	assert.NotEqual(t, errorFingerprint(base), errorFingerprint(
		&hperrors.ErrorInfo{Code: "ERR_PANIC", StackTrace: "main.go:1"}))

	// Detail and cause carry ids and names. Including them would make every
	// occurrence unique and the guard would never fire.
	assert.Equal(t, errorFingerprint(base), errorFingerprint(&hperrors.ErrorInfo{
		Code: "ERR_INTERNAL", StackTrace: "main.go:1",
		Detail: "app 123 failed", Cause: "dial tcp 10.0.0.7: refused",
	}))

	assert.Equal(t, "", errorFingerprint(nil))
}
