package appautoscaleserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging"
)

var t0 = time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)

// A function at Concurrency 10 and a 70 % target keeps 7 calls an instance.
func in(current int, load *logging.InvocationLoad, now time.Time) *input {
	return &input{Current: current, Min: 1, Max: 6, Concurrency: 10, Target: 70, Window: time.Minute,
		ScaleInDelay: 5 * time.Minute, Now: now, Load: load}
}

// busy is a load of n calls in flight on average over the minute.
func busy(n float64) *logging.InvocationLoad {
	return &logging.InvocationLoad{BusyMs: n * 60_000, Calls: int64(n * 100)}
}

func TestTheAverageScalesOutOnTheSecondRunAbove(t *testing.T) {
	d, st := decide(in(1, busy(20), t0), state{})
	assert.Equal(t, 1, d.Desired, "one run above is not a trend")
	assert.Equal(t, 1, st.AboveRuns)

	d, st = decide(in(1, busy(20), t0.Add(15*time.Second)), st)
	assert.Equal(t, 3, d.Desired, "20 in flight at 7 an instance")
	assert.InDelta(t, 20, d.InFlight, 1e-9)
	assert.Equal(t, state{AboveRuns: aboveRunsToScaleOut, LastUpAt: t0.Add(15 * time.Second).Unix()}, st,
		"the count is kept: a load still rising is given more the next run")
}

// Once given, a load still rising is given more every run; a run not above
// starts the count again.
func TestARisingLoadIsGivenMoreEveryRun(t *testing.T) {
	_, st := decide(in(1, busy(10), t0), state{})
	d, st := decide(in(1, busy(10), t0.Add(15*time.Second)), st)
	assert.Equal(t, 2, d.Desired)

	d, st = decide(in(2, busy(20), t0.Add(30*time.Second)), st)
	assert.Equal(t, 3, d.Desired, "still rising: the next run")

	d, st = decide(in(3, busy(20), t0.Add(45*time.Second)), st)
	assert.Equal(t, 3, d.Desired)
	assert.Zero(t, st.AboveRuns, "level: the count starts again")

	d, _ = decide(in(3, busy(30), t0.Add(time.Minute)), st)
	assert.Equal(t, 3, d.Desired, "one run above is not a trend")
}

// burstIn is a function's input with the window's last 15 s read apart.
func burstIn(current int, load *logging.InvocationLoad) *input {
	i := in(current, load, t0)
	i.ShortWindow = 15 * time.Second
	return i
}

// A burst - the window's last 15 s needing twice the replicas or more -
// scales out at once, before the minute's average sees it; less waits for the
// average.
func TestABurstOfCallsScalesOutAtOnce(t *testing.T) {
	// 7 calls in flight over the minute: the one instance; 21 over its last
	// 15 s: 3.
	load := &logging.InvocationLoad{BusyMs: 7 * 60_000, Calls: 100, ShortBusyMs: 21 * 15_000, ShortCalls: 60}
	d, st := decide(burstIn(1, load), state{})
	assert.Equal(t, 3, d.Desired)
	assert.Equal(t, "21.0 calls in flight over the last 15s, 7 an instance", d.Reason)
	assert.Equal(t, aboveRunsToScaleOut, st.AboveRuns, "a trend already")

	d, _ = decide(burstIn(2, load), state{})
	assert.Equal(t, 2, d.Desired, "21 is under twice the 2 instances' 14")

	load.ShortBusyMs = 13 * 15_000
	d, st = decide(burstIn(1, load), state{})
	assert.Equal(t, 1, d.Desired, "13 is under twice the instance's 7")
	assert.Zero(t, st.AboveRuns)

	// The larger of the two: the minute's 30 over the last 15 s's 14.
	load = &logging.InvocationLoad{BusyMs: 30 * 60_000, Calls: 100, ShortBusyMs: 14 * 15_000, ShortCalls: 20}
	d, _ = decide(burstIn(1, load), state{})
	assert.Equal(t, 5, d.Desired)
	assert.Equal(t, "30.0 calls in flight, 7.0 an instance", d.Reason)

	// Calls turned away too: the larger.
	load = &logging.InvocationLoad{BusyMs: 7 * 60_000, Calls: 100, Throttled: 1, ShortBusyMs: 28 * 15_000}
	d, _ = decide(burstIn(1, load), state{})
	assert.Equal(t, 4, d.Desired, "the burst's 4 over one more")
}

// An app's burst of requests is acted on at once, before a larger ask from
// CPU, which is given on the runs after. CPU itself never bursts: it stops at
// its limit, and a starting container's would read as one.
func TestAskAppTakesABurstFirst(t *testing.T) {
	in := &appInput{Current: 2, Window: time.Minute, ShortWindow: 15 * time.Second, RequestsTarget: 10,
		RequestsRead: true, Requests: &logging.RequestLoad{BusyMs: 20 * 60_000, ShortBusyMs: 45 * 15_000},
		CPUTarget: 25, CPURead: true, CPU: []*logging.ContainerCPU{{CPU: 1, Limit: 1}, {CPU: 1, Limit: 1}}}
	a, inFlight, ok := askApp(in)
	assert.True(t, ok)
	assert.True(t, a.Urgent)
	assert.Equal(t, 5, a.Want, "45 in flight over the last 15 s, before CPU's 8")
	assert.Equal(t, "45.0 requests in flight over the last 15s, 10 an instance", a.Up)
	assert.Equal(t, "20.0 requests in flight, CPU at 100 %", a.Low)
	assert.InDelta(t, 20, inFlight, 1e-9, "the minute's, shown")

	in.Requests.ShortBusyMs = 35 * 15_000
	a, _, _ = askApp(in)
	assert.False(t, a.Urgent, "35 is under twice the 2 instances' 20")
	assert.Equal(t, 8, a.Want, "CPU's")
}

// A run grows the replicas twice over or by 4, the larger, at most: a burst
// read over 15 s, or a minute that kept asking, does not overshoot at once.
func TestARunGrowsTwiceOverOrByFourAtMost(t *testing.T) {
	b := bounds{Current: 2, Min: 1, Max: 50, ScaleInDelay: 5 * time.Minute, Now: t0}
	burst := ask{Want: 40, Urgent: true, Up: "a burst"}
	d, _ := settle(b, burst, state{})
	assert.Equal(t, 6, d.Desired, "4 more")
	assert.Equal(t, "a burst; 6 at most this run", d.Reason)

	b.Current = 10
	d, _ = settle(b, burst, state{})
	assert.Equal(t, 20, d.Desired, "twice over")

	d, _ = settle(b, ask{Want: 40, Up: "the average"}, state{AboveRuns: 1})
	assert.Equal(t, 20, d.Desired, "the average too")

	d, _ = settle(b, ask{Want: 15, Urgent: true, Up: "a burst"}, state{})
	assert.Equal(t, 15, d.Desired, "within it, what is asked")
	assert.Equal(t, "a burst", d.Reason)
}

func TestThrottledCallsScaleOutAtOnce(t *testing.T) {
	d, _ := decide(in(2, &logging.InvocationLoad{BusyMs: 60_000, Calls: 10, Throttled: 10}, t0), state{})
	assert.Equal(t, 4, d.Desired, "half the calls turned away: doubled, no more")
	assert.Contains(t, d.Reason, "10 of 10 calls turned away")

	d, _ = decide(in(5, &logging.InvocationLoad{Calls: 100, Throttled: 1}, t0), state{})
	assert.Equal(t, 6, d.Desired, "one turned away: one more")

	d, _ = decide(in(6, &logging.InvocationLoad{Calls: 100, Throttled: 50}, t0), state{})
	assert.Equal(t, 6, d.Desired, "never past Max")
}

func TestItScalesInSlowlyAfterTheDelay(t *testing.T) {
	d, st := decide(in(6, nil, t0), state{})
	assert.Equal(t, 6, d.Desired, "low starts the clock")
	assert.Equal(t, t0.Unix(), st.BelowSince)

	d, st = decide(in(6, nil, t0.Add(4*time.Minute)), st)
	assert.Equal(t, 6, d.Desired, "not yet")

	d, st = decide(in(6, nil, t0.Add(5*time.Minute)), st)
	assert.Equal(t, 3, d.Desired, "half the gap to 1, rounded up")

	d, _ = decide(in(3, nil, t0.Add(5*time.Minute+15*time.Second)), st)
	assert.Equal(t, 2, d.Desired, "the next run, the next half")

	// Load back above resets the clock.
	_, st = decide(in(6, nil, t0), state{})
	_, st = decide(in(6, busy(50), t0.Add(time.Minute)), st)
	assert.Zero(t, st.BelowSince)
}

func TestBoundsComeFirst(t *testing.T) {
	d, _ := decide(in(0, nil, t0), state{})
	assert.Equal(t, 1, d.Desired, "below Min")
	d, _ = decide(in(9, busy(100), t0), state{})
	assert.Equal(t, 6, d.Desired, "above Max")
}

func TestASteadyLoadChangesNothing(t *testing.T) {
	d, st := decide(in(3, busy(18), t0), state{AboveRuns: 1, BelowSince: 5})
	assert.Equal(t, 3, d.Desired)
	assert.Empty(t, d.Reason)
	assert.Equal(t, state{}, st)
}

// The cooldown holds a scale-out it applies to; an urgent one does not wait.
func TestSettleCooldown(t *testing.T) {
	b := bounds{Current: 2, Min: 1, Max: 10, ScaleInDelay: 5 * time.Minute, Now: t0}
	up := ask{Want: 4, Up: "busy", Cooldown: time.Minute}
	st := state{AboveRuns: 1, LastUpAt: t0.Add(-30 * time.Second).Unix()}

	d, next := settle(b, up, st)
	assert.Equal(t, 2, d.Desired, "30 s after the last scale-out")
	assert.Equal(t, 2, next.AboveRuns)

	b.Now = t0.Add(30 * time.Second)
	d, _ = settle(b, up, next)
	assert.Equal(t, 4, d.Desired, "a minute after")

	up.Urgent = true
	d, _ = settle(bounds{Current: 2, Min: 1, Max: 10, Now: t0}, up, state{LastUpAt: t0.Unix()})
	assert.Equal(t, 4, d.Desired)
}

// An app's signals: the larger wins; one that cannot be read is left out;
// none read, it holds.
func TestAskApp(t *testing.T) {
	in := &appInput{Current: 2, Window: time.Minute, RequestsTarget: 10, RequestsRead: true,
		Requests: &logging.RequestLoad{BusyMs: 25 * 60_000}, CPUTarget: 50, CPURead: true,
		CPU: []*logging.ContainerCPU{{CPU: 0.9, Limit: 1}, {CPU: 0.9, Limit: 1}}}
	a, inFlight, ok := askApp(in)
	assert.True(t, ok)
	assert.InDelta(t, 25, inFlight, 1e-9)
	assert.Equal(t, 4, a.Want, "CPU's 2 * 1.8 over the requests' 3")
	assert.Contains(t, a.Up, "CPU at 90 %")
	assert.Equal(t, "25.0 requests in flight, CPU at 90 %", a.Low)

	in.CPURead = false
	a, _, ok = askApp(in)
	assert.True(t, ok)
	assert.Equal(t, 3, a.Want)

	in.RequestsRead = false
	_, _, ok = askApp(in)
	assert.False(t, ok)

	// No limit: measured by the reservation; neither, not measured.
	in.CPURead, in.RequestsTarget = true, 0
	in.CPU = []*logging.ContainerCPU{{CPU: 0.25}, {CPU: 0.25}}
	in.Reservation = 0.25
	a, _, ok = askApp(in)
	assert.True(t, ok)
	assert.Equal(t, 4, a.Want, "at its reservation, twice the target")
	in.Reservation = 0
	_, _, ok = askApp(in)
	assert.False(t, ok)
}
