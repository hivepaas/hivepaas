package tasklog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// fakeRedis keeps what a remote store pushes, deletes and publishes: a list
// per key, and the messages sent on its channel.
type fakeRedis struct {
	redis.UniversalClient
	lists     map[string][]any
	published []string
}

func newFakeRedis() *fakeRedis { return &fakeRedis{lists: map[string][]any{}} }

func (f *fakeRedis) RPush(_ context.Context, key string, values ...any) *redis.IntCmd {
	f.lists[key] = append(f.lists[key], values...)
	return redis.NewIntResult(int64(len(f.lists[key])), nil)
}

func (f *fakeRedis) Expire(context.Context, string, time.Duration) *redis.BoolCmd {
	return redis.NewBoolResult(true, nil)
}

func (f *fakeRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	for _, key := range keys {
		delete(f.lists, key)
	}
	return redis.NewIntResult(int64(len(keys)), nil)
}

func (f *fakeRedis) Publish(_ context.Context, _ string, message any) *redis.IntCmd {
	f.published = append(f.published, fmt.Sprint(message))
	return redis.NewIntResult(1, nil)
}

// Rotating hands what was logged on and starts the list again, telling those
// following to start over at the new one - not that the log has ended.
func TestStore_RotateKeepsFollowersFollowing(t *testing.T) {
	r := newFakeRedis()
	store := NewRemoteStore("task:t1:log", r)
	assert.NoError(t, store.Add(context.Background(), NewOutFrame("step 1", TsNow)))

	assert.NoError(t, store.Rotate())

	assert.Empty(t, r.lists["task:t1:log"])
	assert.Equal(t, []string{string(CommandNewData), string(CommandReset)}, r.published)
	assert.Empty(t, store.frames)
}

// A flush past the size limit starts the list again too: those following are
// told to start over, or they would wait for frames past what the new list has.
func TestStore_FlushTellsFollowersToStartOver(t *testing.T) {
	r := newFakeRedis()
	store := NewRemoteStore("task:t1:log", r)
	store.SetOnFlush(10, func(context.Context, []*LogFrame) error { return nil })

	assert.NoError(t, store.Add(context.Background(), NewOutFrame("hello", TsNow), NewOutFrame("world", TsNow)))

	assert.Empty(t, r.lists["task:t1:log"])
	assert.Equal(t, []string{string(CommandNewData), string(CommandReset)}, r.published)
}

// A follower reads on from where it is on new data, from the start of the new
// list on a reset, and stops when the log is closed.
func TestFollowingTheCommands(t *testing.T) {
	index := int64(7)
	fetch, stop := follow(CommandNewData, &index)
	assert.True(t, fetch)
	assert.False(t, stop)
	assert.Equal(t, int64(7), index)

	fetch, stop = follow(CommandReset, &index)
	assert.True(t, fetch)
	assert.False(t, stop)
	assert.Equal(t, int64(0), index)

	_, stop = follow(CommandClosed, &index)
	assert.True(t, stop)
}
