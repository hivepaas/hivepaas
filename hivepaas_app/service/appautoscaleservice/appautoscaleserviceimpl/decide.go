package appautoscaleserviceimpl

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/services/logging"
)

// aboveRunsToScaleOut is how many runs in a row the load must ask for more
// before it is given: a burst within one run is not a trend. Throttled calls
// do not wait.
const aboveRunsToScaleOut = 2

// maxGrowth bounds a run's scale-out for throttled calls: doubling at most.
const maxGrowth = 2

// cpuTolerance is how far from its target an app's CPU may be and change
// nothing: CPU moves a little every run, and a replica each way every minute
// is noise.
const cpuTolerance = 0.1

// cpuCooldown is how long a scale-out from CPU waits after the last one: a
// starting container burns CPU, and would ask for more.
const cpuCooldown = time.Minute

// bounds are what every decision is made within: the replicas now, Min and
// Max, the scale-in delay, the time.
type bounds struct {
	Current      int
	Min, Max     int
	ScaleInDelay time.Duration
	Now          time.Time
}

// ask is what one run's load asks of an app's replicas.
type ask struct {
	// Want is the replicas the load needs, before Min and Max.
	Want int
	// Urgent scales out at once: calls are being turned away.
	Urgent bool
	// Up says why it scales out; Low what the load is, when it scales in.
	Up, Low string
	// Cooldown is how long a scale-out waits after the last one; 0 for none.
	Cooldown time.Duration
}

// state is what a decision carries to the next run of the same app.
type state struct {
	AboveRuns  int   `json:"aboveRuns,omitempty"`
	BelowSince int64 `json:"belowSince,omitempty"`
	// LastUpAt is when it last scaled out, for the cooldown.
	LastUpAt int64 `json:"lastUpAt,omitempty"`
	// ShortSince is since when its service has run fewer tasks than it
	// wants: the cluster cannot place them.
	ShortSince int64 `json:"shortSince,omitempty"`
}

// decision is an app's replicas after a run, and why.
type decision struct {
	Desired  int
	InFlight float64
	Reason   string
}

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

// decide is a function's replicas for this run, from its calls.
func decide(in *input, st state) (decision, state) {
	a, inFlight := askFunction(in)
	d, next := settle(bounds{Current: in.Current, Min: in.Min, Max: in.Max, ScaleInDelay: in.ScaleInDelay,
		Now: in.Now}, a, st)
	d.InFlight = inFlight
	return d, next
}

// askFunction is what a function's calls ask: as many instances as its calls
// in flight need at Target of its Concurrency; more at once for calls turned
// away, in proportion to how many were, never more than doubling in a run.
func askFunction(in *input) (ask, float64) {
	var inFlight float64
	var calls, throttled int64
	if in.Load != nil && in.Window > 0 {
		inFlight = in.Load.BusyMs / float64(in.Window.Milliseconds())
		calls, throttled = in.Load.Calls, in.Load.Throttled
	}
	perInstance := float64(max(in.Concurrency, 1)) * float64(in.Target) / 100 //nolint:mnd // percent
	a := ask{
		Want: int(math.Ceil(inFlight / perInstance)),
		Up:   fmt.Sprintf("%.1f calls in flight, %.1f an instance", inFlight, perInstance),
		Low:  fmt.Sprintf("%.1f calls in flight", inFlight),
	}
	if throttled > 0 {
		grown := int(math.Ceil(float64(in.Current) * (1 + float64(throttled)/float64(max(calls, 1)))))
		a.Want = max(a.Want, in.Current+1, min(grown, in.Current*maxGrowth))
		a.Urgent = true
		a.Up = fmt.Sprintf("%d of %d calls turned away", throttled, calls)
	}
	return a, inFlight
}

// appInput is what an app's decision is made from: the signals it scales on,
// each with whether it could be read this run.
type appInput struct {
	Current int
	Window  time.Duration

	// RequestsTarget is the requests in flight an instance takes; 0 when the
	// app does not scale on them. Requests is nil when none ended in the
	// window, the access log having been read.
	RequestsTarget int
	RequestsRead   bool
	Requests       *logging.RequestLoad

	// CPUTarget is the share of an instance's CPU kept busy, in percent; 0
	// when the app does not scale on it. Reservation is an instance's CPU
	// reservation, in cores: what a container with no limit is measured by.
	CPUTarget   int
	CPURead     bool
	CPU         []*logging.ContainerCPU
	Reservation float64
}

// askApp is what an app's signals ask; the larger wins. false when none of
// those it scales on could be read: it holds.
func askApp(in *appInput) (ask, float64, bool) {
	var asks []ask
	var inFlight float64
	if in.RequestsTarget > 0 && in.RequestsRead {
		var a ask
		a, inFlight = askRequests(in)
		asks = append(asks, a)
	}
	if in.CPUTarget > 0 && in.CPURead {
		if a, ok := askCPU(in); ok {
			asks = append(asks, a)
		}
	}
	if len(asks) == 0 {
		return ask{}, 0, false
	}
	out := asks[0]
	lows := []string{asks[0].Low}
	for _, a := range asks[1:] {
		lows = append(lows, a.Low)
		if a.Want > out.Want {
			out.Want, out.Up, out.Cooldown = a.Want, a.Up, a.Cooldown
		}
	}
	out.Low = strings.Join(lows, ", ")
	return out, inFlight, true
}

// askRequests: as many instances as the requests in flight need, by Little's
// law over the window.
func askRequests(in *appInput) (ask, float64) {
	var inFlight float64
	if in.Requests != nil && in.Window > 0 {
		inFlight = in.Requests.BusyMs / float64(in.Window.Milliseconds())
	}
	return ask{
		Want: int(math.Ceil(inFlight / float64(in.RequestsTarget))),
		Up:   fmt.Sprintf("%.1f requests in flight, %d an instance", inFlight, in.RequestsTarget),
		Low:  fmt.Sprintf("%.1f requests in flight", inFlight),
	}, inFlight
}

// askCPU: the HPA's formula, the replicas times how far CPU is from its
// target; within the tolerance, as it is. false when no container could be
// measured: none with a limit, and no reservation.
func askCPU(in *appInput) (ask, bool) {
	utilization, ok := cpuUtilization(in.CPU, in.Reservation)
	if !ok {
		return ask{}, false
	}
	target := float64(in.CPUTarget) / 100 //nolint:mnd // percent
	ratio := utilization / target
	want := in.Current
	if math.Abs(ratio-1) > cpuTolerance {
		want = int(math.Ceil(float64(in.Current) * ratio))
	}
	return ask{
		Want:     want,
		Up:       fmt.Sprintf("CPU at %.0f %%, %d %% wanted", utilization*100, in.CPUTarget), //nolint:mnd // percent
		Low:      fmt.Sprintf("CPU at %.0f %%", utilization*100),                             //nolint:mnd // percent
		Cooldown: cpuCooldown,
	}, true
}

// cpuUtilization is the app's containers' CPU over their limits - or, for one
// with none, the reservation - as a share. false when none could be measured.
func cpuUtilization(containers []*logging.ContainerCPU, reservation float64) (float64, bool) {
	var used, capacity float64
	for _, c := range containers {
		limit := c.Limit
		if limit <= 0 {
			limit = reservation
		}
		if limit <= 0 {
			continue
		}
		used += c.CPU
		capacity += limit
	}
	if capacity == 0 {
		return 0, false
	}
	return used / capacity, true
}

// settle is the replicas a run's ask comes to: within the bounds; up fast -
// at once when urgent, else after aboveRunsToScaleOut runs and the cooldown;
// down slow - once below for the scale-in delay, by half the gap a run.
func settle(b bounds, a ask, st state) (decision, state) {
	d := decision{Desired: b.Current}
	want := min(max(a.Want, b.Min), b.Max)
	now := b.Now.Unix()

	switch {
	case b.Current < b.Min || b.Current > b.Max:
		d.Desired, d.Reason = min(max(b.Current, b.Min), b.Max), "outside its bounds"
		return d, state{LastUpAt: st.LastUpAt, ShortSince: st.ShortSince}
	case want > b.Current && a.Urgent:
		d.Desired, d.Reason = want, a.Up
		return d, state{LastUpAt: now, ShortSince: st.ShortSince}
	case want > b.Current:
		st.BelowSince = 0
		st.AboveRuns++
		if st.AboveRuns < aboveRunsToScaleOut {
			return d, st
		}
		if a.Cooldown > 0 && st.LastUpAt > 0 && b.Now.Sub(time.Unix(st.LastUpAt, 0)) < a.Cooldown {
			return d, st
		}
		d.Desired, d.Reason = want, a.Up
		return d, state{LastUpAt: now, ShortSince: st.ShortSince}
	case want < b.Current:
		st.AboveRuns = 0
		if st.BelowSince == 0 {
			st.BelowSince = now
			return d, st
		}
		if b.Now.Sub(time.Unix(st.BelowSince, 0)) < b.ScaleInDelay {
			return d, st
		}
		// Half the gap a run, at least one: a quiet minute does not undo a
		// busy hour, and the clock keeps running for the next step down.
		d.Desired = b.Current - max((b.Current-want+1)/2, 1) //nolint:mnd // half
		d.Reason = fmt.Sprintf("%s, low for %s", a.Low, b.ScaleInDelay)
		return d, st
	}
	return d, state{LastUpAt: st.LastUpAt, ShortSince: st.ShortSince}
}
