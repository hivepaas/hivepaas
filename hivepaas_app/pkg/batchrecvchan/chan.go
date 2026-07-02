package batchrecvchan

import (
	"sync"
	"time"

	"github.com/tiendc/gofn"
)

const (
	minThresholdPeriod = 50 * time.Millisecond
	defaultChanSize    = 100
)

// Options sends data when either the period of time passes or the item count
// reaches the max value.
type Options struct {
	ThresholdPeriod time.Duration
	MaxItem         int
	ChanSize        int // 100 if unset
}

type Chan[T any] struct {
	thresholdPeriod time.Duration
	maxItem         int

	channel      chan []T
	mu           sync.Mutex
	currentBatch []T
	batchMode    bool
	stopped      bool
}

func (ch *Chan[T]) Send(items ...T) {
	ch.mu.Lock()
	if ch.stopped {
		ch.mu.Unlock()
		return
	}

	if !ch.batchMode {
		ch.channel <- items
		ch.mu.Unlock()
		return
	}

	ch.currentBatch = append(ch.currentBatch, items...)
	sendData := len(ch.currentBatch) >= ch.maxItem
	ch.mu.Unlock()

	if sendData {
		ch.sendData()
	}
}

func (ch *Chan[T]) sendData() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.stopped {
		return
	}

	if len(ch.currentBatch) > 0 {
		ch.channel <- ch.currentBatch
		// Allocate a new slice so we don't corrupt data being read by the receiver
		ch.currentBatch = make([]T, 0, ch.maxItem)
	}
}

func (ch *Chan[T]) Receiver() <-chan []T {
	return ch.channel
}

func (ch *Chan[T]) CloseFunc() func() error {
	return func() error { return ch.Close() }
}

func (ch *Chan[T]) Close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.stopped {
		return nil
	}

	// Flush any remaining items before closing
	if ch.batchMode && len(ch.currentBatch) > 0 {
		ch.channel <- ch.currentBatch
		ch.currentBatch = nil
	}

	ch.stopped = true
	close(ch.channel)
	return nil
}

func NewChan[T any](options Options) *Chan[T] {
	ch := &Chan[T]{
		thresholdPeriod: options.ThresholdPeriod,
		maxItem:         options.MaxItem,
	}
	if ch.thresholdPeriod > 0 && ch.thresholdPeriod < minThresholdPeriod {
		ch.thresholdPeriod = minThresholdPeriod
	}
	if ch.thresholdPeriod == 0 && ch.maxItem == 0 {
		ch.maxItem = 1
	}
	ch.batchMode = ch.thresholdPeriod > 0 || ch.maxItem > 1
	if ch.batchMode {
		ch.currentBatch = make([]T, 0, ch.maxItem)
	}
	ch.channel = make(chan []T, gofn.Coalesce(options.ChanSize, defaultChanSize))

	if ch.thresholdPeriod > 0 {
		go func() {
			// Use NewTicker instead of time.Tick to prevent memory leaks
			ticker := time.NewTicker(ch.thresholdPeriod)
			defer ticker.Stop()

			for range ticker.C {
				ch.mu.Lock()
				isStopped := ch.stopped
				ch.mu.Unlock()

				if isStopped {
					return
				}

				ch.sendData()
			}
		}()
	}
	return ch
}
