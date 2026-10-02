package functionautoscaleserviceimpl

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
	assert.Equal(t, state{}, st)
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
