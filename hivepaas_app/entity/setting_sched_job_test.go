package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func inZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
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

	vn := inZone(t, "Asia/Ho_Chi_Minh")
	runs, err = s.CalcNextRuns(from, 2)
	assert.NoError(t, err)
	// 02:00 at UTC+7: 19:00 UTC the day before. 2 Oct 02:00 there is the first
	// after 1 Oct 07:00 there, the initial time.
	assert.Equal(t, []time.Time{time.Date(2026, 10, 1, 19, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 2, 19, 0, 0, 0, time.UTC)}, inUTC(runs))
	assert.Equal(t, vn, runs[0].Location(), "in the installation's timezone")

	inRange, err := s.CalcNextRunsInRange(from, from.Add(48*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, runs, inRange)
}

// An initial time given with an offset of its own - by an API client - has the
// expression read at that offset, whatever the installation's timezone, and
// the runs come back at it, as before there was one.
func TestCronKeepsTheOffsetItWasGiven(t *testing.T) {
	inZone(t, "Asia/Ho_Chi_Minh")
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
	inZone(t, "America/New_York")
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
	s := &SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day), InitialTime: from.Add(17 * time.Hour)}

	inZone(t, "Asia/Ho_Chi_Minh")
	runs, err := s.CalcNextRuns(from, 2)
	assert.NoError(t, err)
	assert.Equal(t, []time.Time{from.Add(17 * time.Hour), from.Add(41 * time.Hour)}, runs)
}
