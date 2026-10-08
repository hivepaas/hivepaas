package queueimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// A task that has not started is looked for over the last hour, not only since
// the last scan: one missed while the scheduler did not run - stopped by an
// update, or a process that was restarting - was otherwise never run at all.
func TestMissedTasksAreLookedForOverTheLastHour(t *testing.T) {
	now := time.Date(2026, 10, 8, 7, 20, 0, 0, time.UTC)

	from, to := scanWindow(now, 10*time.Minute)

	assert.Equal(t, now.Add(-time.Hour), from)
	assert.Equal(t, now.Add(10*time.Minute), to)
}

// Scans further apart than an hour still look back to the one before.
func TestALongCheckIntervalLooksBackToTheLastScan(t *testing.T) {
	now := time.Date(2026, 10, 8, 7, 20, 0, 0, time.UTC)

	from, _ := scanWindow(now, 2*time.Hour)

	assert.Equal(t, now.Add(-2*time.Hour), from)
}
