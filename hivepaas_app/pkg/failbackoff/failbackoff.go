// Package failbackoff spaces out repeated failures of the same check by the same
// actor, so guessing a secret costs time that grows with every wrong answer.
//
// It counts per actor rather than per source address on purpose. The checks it
// guards sit behind a session that is already authenticated, so the caller can
// change address freely; what they cannot change is who they are logged in as.
// A network-level rate limit is a different tool for a different attack and does
// not replace this one.
package failbackoff

import "time"

// maxExponent caps the doubling. Left alone the shift overflows, and long before
// that the wait outlives the store the attempt is kept in, so anything past this
// is arithmetic nobody observes.
const maxExponent = 7

// Attempt is the running count of consecutive failures for one actor.
type Attempt struct {
	Fails       int       `json:"fails"`
	FirstFailAt time.Time `json:"firstFailAt"`
	LastFailAt  time.Time `json:"lastFailAt"`
}

// Policy is how hard the spacing bites.
type Policy struct {
	// MaxFailsInARow is how many failures are free before any waiting starts, and
	// also the size of each step after that: every further MaxFailsInARow doubles
	// the wait.
	MaxFailsInARow int

	// Step is the wait after the first threshold is crossed.
	Step time.Duration
}

// Wait reports how long the actor must wait before the next attempt is allowed.
// Zero means go ahead.
//
// The wait is measured from the *last* failure, not the first. Measured from the
// first, an attacker who spaces guesses out never waits at all: the window is
// already spent by the time they reach the threshold, so the only thing that gets
// stopped is a burst. Measured from the last, slowing down does not buy anything
// back.
func (p Policy) Wait(attempt *Attempt, timeNow time.Time) time.Duration {
	// An unset policy blocks nobody. Both callers build theirs from constants, so
	// this is a guard against a zero value, not a supported configuration.
	if attempt == nil || p.MaxFailsInARow <= 0 || attempt.Fails < p.MaxFailsInARow {
		return 0
	}

	exponent := attempt.Fails / p.MaxFailsInARow
	if exponent > maxExponent {
		exponent = maxExponent
	}
	minWait := time.Duration(1<<exponent) * p.Step

	elapsed := timeNow.Sub(lastFailOf(attempt))
	if elapsed >= minWait {
		return 0
	}
	return minWait - elapsed
}

// lastFailOf is when the run of failures was last added to.
//
// It falls back to the first failure because an attempt may have been written by
// an older build that only kept that one. Treating the missing timestamp as the
// zero time instead would read as "the wait elapsed long ago" and let everybody
// currently being slowed down through at once, on the deploy that introduced it.
func lastFailOf(attempt *Attempt) time.Time {
	if attempt.LastFailAt.IsZero() {
		return attempt.FirstFailAt
	}
	return attempt.LastFailAt
}

// Fail records one more failure and returns the updated attempt. A nil attempt
// starts a new count, so the caller does not have to distinguish the first
// failure from the rest.
func (p Policy) Fail(attempt *Attempt, timeNow time.Time) *Attempt {
	if attempt == nil {
		attempt = &Attempt{}
	}
	attempt.Fails++
	if attempt.FirstFailAt.IsZero() {
		attempt.FirstFailAt = timeNow
	}
	attempt.LastFailAt = timeNow
	return attempt
}
