package failbackoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testPolicy = Policy{MaxFailsInARow: 5, Step: 2 * time.Minute}

func TestPolicyWait(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)

	t.Run("no attempt yet", func(t *testing.T) {
		assert.Zero(t, testPolicy.Wait(nil, now))
	})

	t.Run("below the threshold", func(t *testing.T) {
		attempt := &Attempt{Fails: 4, LastFailAt: now}
		assert.Zero(t, testPolicy.Wait(attempt, now))
	})

	t.Run("doubles at each further threshold", func(t *testing.T) {
		for _, tc := range []struct {
			fails int
			want  time.Duration
		}{
			{5, 4 * time.Minute},
			{9, 4 * time.Minute},
			{10, 8 * time.Minute},
			{15, 16 * time.Minute},
			{20, 32 * time.Minute},
		} {
			attempt := &Attempt{Fails: tc.fails, LastFailAt: now}
			assert.Equal(t, tc.want, testPolicy.Wait(attempt, now), "fails=%d", tc.fails)
		}
	})

	t.Run("counts down as time passes", func(t *testing.T) {
		attempt := &Attempt{Fails: 5, LastFailAt: now.Add(-3 * time.Minute)}
		assert.Equal(t, time.Minute, testPolicy.Wait(attempt, now))

		attempt.LastFailAt = now.Add(-4 * time.Minute)
		assert.Zero(t, testPolicy.Wait(attempt, now))
	})

	// The whole point of measuring from the last failure. Spread the guesses out
	// and, measured from the first, the window is already spent before the
	// threshold is even reached - the attacker never waits.
	t.Run("slowing down does not buy the wait back", func(t *testing.T) {
		firstFail := now.Add(-time.Hour)
		attempt := &Attempt{Fails: 5, FirstFailAt: firstFail, LastFailAt: now}
		assert.Equal(t, 4*time.Minute, testPolicy.Wait(attempt, now))
	})

	// Unbounded doubling would overflow the shift.
	t.Run("the exponent is capped", func(t *testing.T) {
		attempt := &Attempt{Fails: 5 * 1000, LastFailAt: now}
		assert.Equal(t, time.Duration(1<<maxExponent)*testPolicy.Step,
			testPolicy.Wait(attempt, now))
	})

	t.Run("a zero policy blocks nobody", func(t *testing.T) {
		assert.Zero(t, Policy{}.Wait(&Attempt{Fails: 1000, LastFailAt: now}, now))
	})
}

func TestPolicyFail(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)

	t.Run("starts a count", func(t *testing.T) {
		attempt := testPolicy.Fail(nil, now)
		assert.Equal(t, 1, attempt.Fails)
		assert.Equal(t, now, attempt.FirstFailAt)
		assert.Equal(t, now, attempt.LastFailAt)
	})

	t.Run("keeps the first timestamp and moves the last", func(t *testing.T) {
		attempt := testPolicy.Fail(nil, now)
		later := now.Add(time.Minute)
		attempt = testPolicy.Fail(attempt, later)

		assert.Equal(t, 2, attempt.Fails)
		assert.Equal(t, now, attempt.FirstFailAt)
		assert.Equal(t, later, attempt.LastFailAt)
	})
}
