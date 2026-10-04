package cacherepository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Run against a Redis whose URL is in HP_TEST_REDIS_URL: one of its own, the
// test uses the keys it caches with.
func TestLiveOBISettingsCacheTakesTheCurrentGenerationOnly(t *testing.T) {
	url := os.Getenv("HP_TEST_REDIS_URL")
	if url == "" {
		t.Skip("HP_TEST_REDIS_URL not set")
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(opts)
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	client.Del(ctx, obiSettingsKey, obiSettingsGenKey)
	repo := NewOBISettingsRepo(client)

	got, gen, err := repo.Get(ctx)
	assert.NoError(t, err)
	assert.Nil(t, got, "nothing cached")
	assert.Zero(t, gen)

	on := &entity.OBIAgentSettings{LogsOn: true, Performance: &entity.LoggingPerformance{Enabled: true},
		Apps: map[string]string{"p1_dev_a1": "A1"}}
	assert.NoError(t, repo.Set(ctx, on, gen, time.Hour))
	got, _, err = repo.Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, on, got)
	ttl := client.TTL(ctx, obiSettingsKey).Val()
	assert.Greater(t, ttl, 59*time.Minute)

	// A change: the entry goes, the generation moves.
	assert.NoError(t, repo.Invalidate(ctx))
	got, gen2, err := repo.Get(ctx)
	assert.NoError(t, err)
	assert.Nil(t, got)
	assert.Equal(t, gen+1, gen2)

	// An agent that read the database before the change caches what it read
	// after: under the old generation, it is not taken.
	assert.NoError(t, repo.Set(ctx, on, gen, time.Hour))
	got, _, err = repo.Get(ctx)
	assert.NoError(t, err)
	assert.Nil(t, got, "read before the change")

	off := &entity.OBIAgentSettings{LogsOn: true}
	assert.NoError(t, repo.Set(ctx, off, gen2, time.Hour))
	got, _, err = repo.Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, off, got)
	assert.False(t, got.On())

	// One written by another version, unreadable: a miss, not an error.
	client.Set(ctx, obiSettingsKey, "not json", time.Minute)
	got, _, err = repo.Get(ctx)
	assert.NoError(t, err)
	assert.Nil(t, got)

	client.Del(ctx, obiSettingsKey, obiSettingsGenKey)
}
