package timeutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSleepCtx_NormalSleep(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := SleepCtx(ctx, 50*time.Millisecond)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

func TestSleepCtx_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()

	// Cancel context after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Try to sleep for 500ms
	err := SleepCtx(ctx, 500*time.Millisecond)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, duration, 500*time.Millisecond)
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

func TestSleepCtx_AlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	start := time.Now()
	err := SleepCtx(ctx, 500*time.Millisecond)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, duration, 10*time.Millisecond) // Should return instantly
}

func TestSleepCtx_ZeroOrNegativeSleep(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	err := SleepCtx(ctx, 0)
	assert.NoError(t, err)

	err = SleepCtx(ctx, -10*time.Millisecond)
	assert.NoError(t, err)

	assert.Less(t, time.Since(start), 10*time.Millisecond) // Should return instantly
}
