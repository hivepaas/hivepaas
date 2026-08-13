package tasklog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redact"
)

func TestStore_LocalAddAndGet(t *testing.T) {
	store := NewLocalStore("testkey")

	f1 := NewInFrame("msg1", TsNow)
	f2 := NewOutFrame("msg2", TsNow)

	err := store.Add(context.Background(), f1, f2)
	assert.NoError(t, err)

	frames, err := store.GetData(context.Background(), 0)
	assert.NoError(t, err)
	assert.Len(t, frames, 2)
	assert.Equal(t, f1, frames[0])
	assert.Equal(t, f2, frames[1])
}

func TestStore_AddRedacted(t *testing.T) {
	r := redact.New([]string{"secret123"})

	store := NewLocalStore("testkey")
	store.SetRedactor(r)

	f1 := NewInFrame("normal message", TsNow)
	f2 := NewOutFrame("leak secret123 in log", TsNow)

	err := store.AddRedacted(context.Background(), f1, f2)
	assert.NoError(t, err)

	frames, err := store.GetData(context.Background(), 0)
	assert.NoError(t, err)
	assert.Len(t, frames, 2)

	assert.Equal(t, "normal message", frames[0].Data)
	assert.Equal(t, "leak ******** in log", frames[1].Data)
}

func TestStore_FlushThreshold(t *testing.T) {
	store := NewLocalStore("testkey")

	var flushedFrames []*LogFrame
	var callCount int

	store.SetOnFlush(10, func(ctx context.Context, frames []*LogFrame) error {
		flushedFrames = append(flushedFrames, frames...)
		callCount++
		return nil
	})

	// Add a frame of size 5. Size < 10, shouldn't flush.
	err := store.Add(context.Background(), NewOutFrame("hello", TsNow))
	assert.NoError(t, err)
	assert.Equal(t, 0, callCount)
	assert.Equal(t, int64(5), store.totalSize)

	// Add another frame of size 5. Total size becomes 10 >= 10. Should flush!
	err = store.Add(context.Background(), NewOutFrame("world", TsNow))
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Len(t, flushedFrames, 2)
	assert.Equal(t, "hello", flushedFrames[0].Data)
	assert.Equal(t, "world", flushedFrames[1].Data)

	// Store size should be reset
	assert.Equal(t, int64(0), store.totalSize)
	assert.Len(t, store.frames, 0)
}

func TestStore_FlushThreshold_NoStoreLocal(t *testing.T) {
	// Initialize with storeLocal = false, storeRemote = false
	store := newStore("testkey", false, false, nil, nil)

	var flushedFrames []*LogFrame
	var callCount int

	store.SetOnFlush(10, func(ctx context.Context, frames []*LogFrame) error {
		flushedFrames = append(flushedFrames, frames...)
		callCount++
		return nil
	})

	// Add a frame of size 5. Shouldn't flush.
	err := store.Add(context.Background(), NewOutFrame("hello", TsNow))
	assert.NoError(t, err)
	assert.Equal(t, 0, callCount)
	assert.Equal(t, int64(5), store.totalSize)

	// Add another frame of size 5. Total size becomes 10 >= 10. Should flush!
	err = store.Add(context.Background(), NewOutFrame("world", TsNow))
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Len(t, flushedFrames, 2)
	assert.Equal(t, "hello", flushedFrames[0].Data)
	assert.Equal(t, "world", flushedFrames[1].Data)

	// Store size and frames should be reset
	assert.Equal(t, int64(0), store.totalSize)
	assert.Len(t, store.frames, 0)
}

func TestStore_Forward(t *testing.T) {
	var forwardedFrames []*LogFrame
	var callCount int

	forwardFunc := func(ctx context.Context, frames []*LogFrame) error {
		forwardedFrames = append(forwardedFrames, frames...)
		callCount++
		return nil
	}

	store := NewForwardStore("forward-key", forwardFunc)
	assert.Equal(t, "forward-key", store.Key)

	f1 := NewOutFrame("line 1", TsNow)
	f2 := NewErrFrame("line 2", TsNow)

	err := store.Add(context.Background(), f1, f2)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Len(t, forwardedFrames, 2)
	assert.Equal(t, "line 1", forwardedFrames[0].Data)
	assert.Equal(t, "line 2", forwardedFrames[1].Data)

	// Verify that frames are also stored locally
	localFrames, err := store.GetData(context.Background(), 0)
	assert.NoError(t, err)
	assert.Len(t, localFrames, 2)
	assert.Equal(t, f1, localFrames[0])
	assert.Equal(t, f2, localFrames[1])
}

func TestStore_Forward_Redacted(t *testing.T) {
	var forwardedFrames []*LogFrame
	forwardFunc := func(ctx context.Context, frames []*LogFrame) error {
		forwardedFrames = append(forwardedFrames, frames...)
		return nil
	}

	store := NewForwardStore("testkey", forwardFunc)
	store.SetRedactor(redact.New([]string{"password123"}))

	err := store.AddRedacted(context.Background(), NewOutFrame("user password123", TsNow))
	assert.NoError(t, err)
	assert.Len(t, forwardedFrames, 1)
	assert.Equal(t, "user ********", forwardedFrames[0].Data)
}

func TestStore_SetOnForward(t *testing.T) {
	store := NewLocalStore("testkey")

	var forwardedFrames []*LogFrame
	store.SetOnForward(func(ctx context.Context, frames []*LogFrame) error {
		forwardedFrames = append(forwardedFrames, frames...)
		return nil
	})

	err := store.Add(context.Background(), NewOutFrame("forwarded msg", TsNow))
	assert.NoError(t, err)
	assert.Len(t, forwardedFrames, 1)
	assert.Equal(t, "forwarded msg", forwardedFrames[0].Data)
}

func TestStore_Forward_Error(t *testing.T) {
	expectedErr := errors.New("forward error")
	store := NewForwardStore("testkey", func(ctx context.Context, frames []*LogFrame) error {
		return expectedErr
	})

	err := store.Add(context.Background(), NewOutFrame("msg", TsNow))
	assert.ErrorIs(t, err, expectedErr)
}

func TestStore_Forward_EmptyFrames(t *testing.T) {
	var called bool
	store := NewForwardStore("testkey", func(ctx context.Context, frames []*LogFrame) error {
		called = true
		return nil
	})

	err := store.Add(context.Background())
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestStore_Forward_ResetAndClose(t *testing.T) {
	store := NewForwardStore("testkey", func(ctx context.Context, frames []*LogFrame) error {
		return nil
	})

	err := store.Add(context.Background(), NewOutFrame("msg", TsNow))
	assert.NoError(t, err)

	err = store.Reset()
	assert.NoError(t, err)

	frames, err := store.GetData(context.Background(), 0)
	assert.NoError(t, err)
	assert.Empty(t, frames)

	err = store.Close()
	assert.NoError(t, err)
}
