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

// scheduled is a schedule that made its tasks up to lastRun, as the scan does.
func scheduled(s SchedJobSchedule, lastRun time.Time) *SchedJobSchedule {
	s.SetLastSchedTime(lastRun)
	return &s
}

// The runs are counted from the last one a task was made for, not from the
// initial time: a job made a year ago to run each minute would otherwise walk
// the whole year again on each scan, and on each look at its page.
func TestRunsAreCountedFromTheLastOneScheduled(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 30, 0, time.UTC)
	lastRun := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	s := scheduled(SchedJobSchedule{CronExpr: "* * * * *", InitialTime: now.AddDate(-1, 0, 0)}, lastRun)

	assert.Equal(t, lastRun, s.countingFrom(now))
}

// A schedule changed since its last run was scheduled counts from its initial
// time again.
func TestAChangedScheduleCountsFromItsInitialTime(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 30, 0, time.UTC)
	initial := now.AddDate(0, -1, 0)
	for name, change := range map[string]func(s *SchedJobSchedule){
		"cron":         func(s *SchedJobSchedule) { s.CronExpr = "*/5 * * * *" },
		"interval":     func(s *SchedJobSchedule) { s.CronExpr, s.Interval = "", timeutil.Duration(time.Hour) },
		"initial time": func(s *SchedJobSchedule) { s.InitialTime = initial.Add(time.Hour) },
	} {
		s := scheduled(SchedJobSchedule{CronExpr: "* * * * *", InitialTime: initial}, now.Add(-30*time.Second))
		change(s)
		assert.Equal(t, s.InitialTime, s.countingFrom(now), name)
	}
}

// A run whose task is made, still to come, is one of the next runs: the job's
// page lists it, and a job turned off and on again makes it again.
func TestARunScheduledButToComeIsStillNext(t *testing.T) {
	now := time.Date(2026, 10, 10, 1, 55, 0, 0, time.UTC)
	nextRun := time.Date(2026, 10, 10, 2, 0, 0, 0, time.UTC)
	cron := scheduled(SchedJobSchedule{CronExpr: "0 2 * * *", InitialTime: now.AddDate(0, -1, 0)}, nextRun)
	every := scheduled(SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day),
		InitialTime: nextRun.AddDate(0, -1, 0)}, nextRun)

	for name, s := range map[string]*SchedJobSchedule{"cron": cron, "interval": every} {
		runs, err := s.CalcNextRuns(now, 1)
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{nextRun}, inUTC(runs), name)

		runs, err = s.CalcNextRunsInRange(now, now.Add(10*time.Minute))
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{nextRun}, inUTC(runs), name)
	}
}

// Counted from the last run scheduled or from the initial time, the runs are
// the same.
func TestRunsAfterTheLastOneScheduledAreTheSchedulesRuns(t *testing.T) {
	inNewYork(t)
	now := time.Date(2026, 10, 31, 23, 58, 0, 0, time.UTC)
	for _, s := range []SchedJobSchedule{
		{CronExpr: "*/7 2-3 * * *", InitialTime: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Interval: timeutil.Duration(7 * time.Hour), InitialTime: time.Date(2026, 9, 1, 0, 3, 0, 0, time.UTC)},
	} {
		want, err := s.CalcNextRunsInRange(now, now.Add(72*time.Hour))
		assert.NoError(t, err)
		lastRun, err := s.CalcNextRuns(now.Add(-36*time.Hour), 1)
		assert.NoError(t, err)

		got, err := scheduled(s, lastRun[0]).CalcNextRunsInRange(now, now.Add(72*time.Hour))
		assert.NoError(t, err)
		assert.NotEmpty(t, want)
		assert.Equal(t, want, got)
	}
}

// A scan that makes no task keeps the last run it knows of.
func TestAScanWithoutRunsKeepsTheLastOne(t *testing.T) {
	lastRun := time.Date(2026, 10, 10, 2, 0, 0, 0, time.UTC)
	s := scheduled(SchedJobSchedule{CronExpr: "0 2 * * *", InitialTime: lastRun.AddDate(0, -1, 0)}, lastRun)

	assert.False(t, s.SetLastSchedTime(time.Time{}))
	assert.Equal(t, lastRun, s.LastSchedTime)
}
