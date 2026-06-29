package tasklog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/localpaas/localpaas/localpaas_app/pkg/redact"
)

func TestStore_LocalAddAndGet(t *testing.T) {
	store := newStore("testkey", true, false, nil)

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

	store := newStore("testkey", true, false, nil)
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
	store := newStore("testkey", true, false, nil)

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
	store := newStore("testkey", false, false, nil)

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
