package functionautoscaleserviceimpl

import (
	"fmt"
	"math"
	"time"

	"github.com/hivepaas/hivepaas/services/logging"
)

// aboveRunsToScaleOut is how many runs in a row the average must ask for more
// before it is given: a burst within one run is not a trend. Throttled calls
// do not wait.
const aboveRunsToScaleOut = 2

// maxGrowth bounds a run's scale-out for throttled calls: doubling at most.
const maxGrowth = 2

// input is what one function's decision is made from.
type input struct {
	Current      int
	Min, Max     int
	Concurrency  int
	Target       int // percent of Concurrency an instance is kept at
	Window       time.Duration
	ScaleInDelay time.Duration
	Now          time.Time
	// Load is nil when the function had no call in the window: no load, the
	// logs having been read.
	Load *logging.InvocationLoad
}

// state is what a decision carries to the next run of the same function.
type state struct {
	AboveRuns  int   `json:"aboveRuns,omitempty"`
	BelowSince int64 `json:"belowSince,omitempty"`
}

// decision is a function's replicas after a run, and why.
type decision struct {
	Desired  int
	InFlight float64
	Reason   string
}

// decide is a function's replicas for this run. Up fast: at once for throttled
// calls, after aboveRunsToScaleOut runs for the average. Down slow: once the
// load has been below for the scale-in delay, by half the gap a run.
func decide(in *input, st state) (decision, state) {
	var inFlight float64
	var calls, throttled int64
	if in.Load != nil && in.Window > 0 {
		inFlight = in.Load.BusyMs / float64(in.Window.Milliseconds())
		calls, throttled = in.Load.Calls, in.Load.Throttled
	}
	perInstance := float64(max(in.Concurrency, 1)) * float64(in.Target) / 100 //nolint:mnd // percent
	want := int(math.Ceil(inFlight / perInstance))
	if throttled > 0 {
		// Turned away now: more than the average says, in proportion to how
		// many were, and never more than doubling in a run.
		grown := int(math.Ceil(float64(in.Current) * (1 + float64(throttled)/float64(max(calls, 1)))))
		want = max(want, in.Current+1, min(grown, in.Current*maxGrowth))
	}
	want = min(max(want, in.Min), in.Max)
	d := decision{Desired: in.Current, InFlight: inFlight}

	switch {
	case in.Current < in.Min || in.Current > in.Max:
		d.Desired, d.Reason = min(max(in.Current, in.Min), in.Max), "outside its bounds"
		return d, state{}
	case want > in.Current && throttled > 0:
		d.Desired = want
		d.Reason = fmt.Sprintf("%d of %d calls turned away", throttled, calls)
		return d, state{}
	case want > in.Current:
		st.BelowSince = 0
		st.AboveRuns++
		if st.AboveRuns < aboveRunsToScaleOut {
			return d, st
		}
		d.Desired = want
		d.Reason = fmt.Sprintf("%.1f calls in flight, %.1f an instance", inFlight, perInstance)
		return d, state{}
	case want < in.Current:
		st.AboveRuns = 0
		if st.BelowSince == 0 {
			st.BelowSince = in.Now.Unix()
			return d, st
		}
		if in.Now.Sub(time.Unix(st.BelowSince, 0)) < in.ScaleInDelay {
			return d, st
		}
		// Half the gap a run, at least one: a quiet minute does not undo a
		// busy hour, and the clock keeps running for the next step down.
		d.Desired = in.Current - max((in.Current-want+1)/2, 1) //nolint:mnd // half
		d.Reason = fmt.Sprintf("%.1f calls in flight, low for %s", inFlight, in.ScaleInDelay)
		return d, st
	}
	return d, state{}
}
