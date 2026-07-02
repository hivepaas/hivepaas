package executil

import (
	"math"
	"math/rand"
	"time"
)

// RetryDelay computes the delay duration for a given retry attempt.
// It supports linear delay increment (if delayIncr != nil), exponential backoff with jitter
// (if backoffJitter != nil), or a constant delay.
func RetryDelay(retry int, delay time.Duration, delayIncr, backoffJitter *time.Duration) time.Duration {
	// Delay with increment
	if delayIncr != nil {
		if retry > 1 {
			return delay + time.Duration(retry-1)*(*delayIncr)
		}
		return delay
	}

	// Expo backoff delay
	if backoffJitter != nil {
		jitter := time.Duration(0)
		if *backoffJitter > 0 {
			jitter = time.Duration(rand.Int63n(int64(*backoffJitter))) //nolint:gosec
		}
		if retry > 1 {
			return time.Duration(math.Pow(2, float64(retry-1))*float64(delay)) + jitter //nolint:mnd
		}
		return delay + jitter
	}

	return delay
}
