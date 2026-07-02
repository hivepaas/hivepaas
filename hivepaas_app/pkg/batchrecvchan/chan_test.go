package batchrecvchan

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChan_NoBatchMode(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 0,
		MaxItem:         1,
	})
	defer ch.Close()

	ch.Send(1)
	ch.Send(2)

	select {
	case batch := <-ch.Receiver():
		assert.Equal(t, []int{1}, batch)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for receiver")
	}

	select {
	case batch := <-ch.Receiver():
		assert.Equal(t, []int{2}, batch)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for receiver")
	}
}

func TestChan_BatchMode_MaxItem(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 10 * time.Second, // make it large so it doesn't trigger
		MaxItem:         3,
	})
	defer ch.Close()

	ch.Send(1)
	ch.Send(2)

	// Should not have received anything yet
	select {
	case <-ch.Receiver():
		t.Fatal("received prematurely")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	ch.Send(3)

	// Should receive the batch now
	select {
	case batch := <-ch.Receiver():
		assert.Equal(t, []int{1, 2, 3}, batch)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for receiver")
	}
}

func TestChan_BatchMode_ThresholdPeriod(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 50 * time.Millisecond,
		MaxItem:         100,
	})
	defer ch.Close()

	ch.Send(1)
	ch.Send(2)

	// Should receive the batch after period
	select {
	case batch := <-ch.Receiver():
		assert.Equal(t, []int{1, 2}, batch)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for receiver")
	}
}

func TestChan_Close_FlushesRemaining(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 10 * time.Second,
		MaxItem:         10,
	})

	ch.Send(1)
	ch.Send(2)

	// Close should flush the 2 items
	err := ch.Close()
	assert.NoError(t, err)

	select {
	case batch := <-ch.Receiver():
		assert.Equal(t, []int{1, 2}, batch)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for receiver")
	}

	// Channel should be closed
	_, ok := <-ch.Receiver()
	assert.False(t, ok, "channel should be closed")
}

func TestChan_Close_SafeAfterClose(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 10 * time.Second,
		MaxItem:         10,
	})
	err := ch.Close()
	assert.NoError(t, err)

	// Sending after close should not panic
	assert.NotPanics(t, func() {
		ch.Send(1, 2, 3)
	})
}

func TestChan_ConcurrentSends(t *testing.T) {
	ch := NewChan[int](Options{
		ThresholdPeriod: 50 * time.Millisecond,
		MaxItem:         5,
	})
	defer ch.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ch.Send(val)
		}(i)
	}
	wg.Wait()

	totalReceived := 0
	for totalReceived < 100 {
		select {
		case batch := <-ch.Receiver():
			totalReceived += len(batch)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for receiver")
		}
	}
	assert.Equal(t, 100, totalReceived)
}
