package executil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func toPtr(d time.Duration) *time.Duration {
	return &d
}

func TestRetryDelay(t *testing.T) {
	// Case 1: delayIncr != nil (linear increment delay)
	t.Run("LinearIncrement", func(t *testing.T) {
		// retry = 0 -> delay
		d := RetryDelay(0, 1*time.Second, toPtr(500*time.Millisecond), nil)
		assert.Equal(t, 1*time.Second, d)

		// retry = 1 -> delay (first retry uses base delay)
		d = RetryDelay(1, 1*time.Second, toPtr(500*time.Millisecond), nil)
		assert.Equal(t, 1*time.Second, d)

		// retry = 2 -> delay + 1*delayIncr
		d = RetryDelay(2, 1*time.Second, toPtr(500*time.Millisecond), nil)
		assert.Equal(t, 1500*time.Millisecond, d)

		// retry = 3 -> delay + 2*delayIncr
		d = RetryDelay(3, 1*time.Second, toPtr(500*time.Millisecond), nil)
		assert.Equal(t, 2000*time.Millisecond, d)
	})

	// Case 2: delayIncr == nil && backoffJitter == nil (constant delay)
	t.Run("ConstantDelay", func(t *testing.T) {
		d := RetryDelay(3, 1*time.Second, nil, nil)
		assert.Equal(t, 1*time.Second, d)
	})

	// Case 3: backoffJitter != nil (exponential backoff)
	t.Run("ExponentialBackoffNoJitter", func(t *testing.T) {
		// retry = 0 -> base delay
		d := RetryDelay(0, 1*time.Second, nil, toPtr(0))
		assert.Equal(t, 1*time.Second, d)

		// retry = 1 -> base delay (first retry uses base delay)
		d = RetryDelay(1, 1*time.Second, nil, toPtr(0))
		assert.Equal(t, 1*time.Second, d)

		// retry = 2 -> 2^1 * delay = 2s
		d = RetryDelay(2, 1*time.Second, nil, toPtr(0))
		assert.Equal(t, 2*time.Second, d)

		// retry = 3 -> 2^2 * delay = 4s
		d = RetryDelay(3, 1*time.Second, nil, toPtr(0))
		assert.Equal(t, 4*time.Second, d)
	})

	t.Run("ExponentialBackoffWithJitter", func(t *testing.T) {
		// When jitter > 0, the delay should be between 2^(retry-1) * delay and 2^(retry-1) * delay + backoffJitter
		jitter := 100 * time.Millisecond
		d := RetryDelay(2, 1*time.Second, nil, toPtr(jitter))

		minExpected := 2 * time.Second
		maxExpected := 2*time.Second + jitter
		assert.True(t, d >= minExpected && d < maxExpected)
	})
}
