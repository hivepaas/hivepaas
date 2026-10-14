package cacherepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
)

const (
	unusedImagesKeyPrefix = "cleanup:unused-images:"
)

// UnusedImagesRepo keeps, for each node, when the cluster cleanup first found
// each of its images unused: Docker records when an image was built or pulled,
// never when it was last used.
//
// Losing it - Redis is not persisted - only starts every count over: images are
// then removed later than they would have been, never earlier.
type UnusedImagesRepo interface {
	// Get answers when each image of a node was first found unused, by image
	// ID; empty when nothing was recorded.
	Get(ctx context.Context, nodeID string) (map[string]time.Time, error)
	// Set replaces a node's record, kept for ttl.
	Set(ctx context.Context, nodeID string, unusedSince map[string]time.Time, ttl time.Duration) error
}

type unusedImagesRepo struct {
	client rediscache.Client
}

func NewUnusedImagesRepo(client rediscache.Client) UnusedImagesRepo {
	return &unusedImagesRepo{client: client}
}

func (repo *unusedImagesRepo) Get(ctx context.Context, nodeID string) (map[string]time.Time, error) {
	data, err := repo.client.Get(ctx, unusedImagesKeyPrefix+nodeID).Bytes()
	if errors.Is(err, redis.Nil) {
		return map[string]time.Time{}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var unixSince map[string]int64
	if err = json.Unmarshal(data, &unixSince); err != nil {
		// Unreadable - written by another version - as if nothing were recorded.
		return map[string]time.Time{}, nil //nolint:nilerr // counts start over, which only delays removals
	}
	unusedSince := make(map[string]time.Time, len(unixSince))
	for imageID, since := range unixSince {
		unusedSince[imageID] = time.Unix(since, 0).UTC()
	}
	return unusedSince, nil
}

func (repo *unusedImagesRepo) Set(
	ctx context.Context,
	nodeID string,
	unusedSince map[string]time.Time,
	ttl time.Duration,
) error {
	unixSince := make(map[string]int64, len(unusedSince))
	for imageID, since := range unusedSince {
		unixSince[imageID] = since.Unix()
	}
	data, err := json.Marshal(unixSince)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(repo.client.Set(ctx, unusedImagesKeyPrefix+nodeID, data, ttl).Err())
}
