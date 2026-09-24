package dockerproxy

import "testing"

// stop ends the test when the assertion it wraps failed, where going on would
// only report the same failure again.
func stop(t testing.TB, ok bool) {
	t.Helper()
	if !ok {
		t.FailNow()
	}
}
