package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// inNewYork makes New York the installation's timezone: UTC-5, and UTC-4 in
// daylight saving, from March to November.
func inNewYork(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	timeutil.SetLocation(loc)
	t.Cleanup(func() { timeutil.SetLocation(nil) })
	return loc
}

func inUTC(runs []time.Time) []time.Time {
	out := make([]time.Time, 0, len(runs))
	for _, run := range runs {
		out = append(out, run.UTC())
	}
	return out
}

// A cron expression's hours are the installation's, for an initial time in
// UTC - as the dashboard gives it - and the runs come back in its timezone.
func TestCronRunsAtTheInstallationsHours(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	s := &SchedJobSchedule{CronExpr: "0 2 * * *", InitialTime: from}

	runs, err := s.CalcNextRuns(from, 2)
	assert.NoError(t, err)
	assert.Equal(t, []time.Time{time.Date(2026, 10, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 2, 2, 0, 0, 0, time.UTC)}, runs, "UTC until set")

	newYork := inNewYork(t)
	runs, err = s.CalcNextRuns(from, 2)
	assert.NoError(t, err)
	// 02:00 at UTC-4, New York's daylight saving: 06:00 UTC. 1 Oct 02:00 there
	// is the first after 30 Sep 20:00 there, the initial time.
	assert.Equal(t, []time.Time{time.Date(2026, 10, 1, 6, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 2, 6, 0, 0, 0, time.UTC)}, inUTC(runs))
	assert.Equal(t, newYork, runs[0].Location(), "in the installation's timezone")

	inRange, err := s.CalcNextRunsInRange(from, from.Add(48*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, runs, inRange)
}

// An initial time given with an offset of its own - by an API client - has the
// expression read at that offset, whatever the installation's timezone, and
// the runs come back at it, as before there was one.
func TestCronKeepsTheOffsetItWasGiven(t *testing.T) {
	inNewYork(t)
	tokyo := time.FixedZone("", 9*60*60)
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, tokyo)
	s := &SchedJobSchedule{CronExpr: "0 2 * * *", InitialTime: from}

	runs, err := s.CalcNextRuns(from, 1)
	assert.NoError(t, err)
	if assert.Len(t, runs, 1) {
		assert.Equal(t, time.Date(2026, 10, 1, 2, 0, 0, 0, tokyo), runs[0])
	}
}

// Across a change of the clocks, a cron expression keeps its hour there.
func TestCronKeepsItsHourAcrossDST(t *testing.T) {
	inNewYork(t)
	from := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	s := &SchedJobSchedule{CronExpr: "0 3 * * *", InitialTime: from}

	runs, err := s.CalcNextRuns(from, 3)
	assert.NoError(t, err)
	// 1 November 2026 the clocks go back: EDT is UTC-4, EST UTC-5.
	assert.Equal(t, []time.Time{
		time.Date(2026, 10, 31, 7, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 2, 8, 0, 0, 0, time.UTC),
	}, inUTC(runs))
}

// An interval is from an instant: the timezone does not move it.
func TestIntervalIsNotMovedByTheTimezone(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	// Midnight in New York, at UTC-4.
	s := &SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day), InitialTime: from.Add(4 * time.Hour)}

	inNewYork(t)
	runs, err := s.CalcNextRuns(from, 2)
	assert.NoError(t, err)
	assert.Equal(t, []time.Time{from.Add(4 * time.Hour), from.Add(28 * time.Hour)}, runs)
}
