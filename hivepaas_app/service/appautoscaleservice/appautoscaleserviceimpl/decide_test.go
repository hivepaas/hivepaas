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
	assert.Equal(t, state{LastUpAt: t0.Add(15 * time.Second).Unix()}, st, "the count starts again")
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
