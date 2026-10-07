package redishelper

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func ZAdd(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	members ...redis.Z,
) error {
	if len(members) == 0 {
		return nil
	}
	_, err := cmder.ZAdd(ctx, key, members...).Result()
	if err != nil {
		return hperrors.Wrap(err).WithMsgLog("failed to add members to zset")
	}
	return nil
}

// ZAddNX adds the members that are not in the set; one already there keeps its
// score.
func ZAddNX(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	members ...redis.Z,
) error {
	if len(members) == 0 {
		return nil
	}
	_, err := cmder.ZAddNX(ctx, key, members...).Result()
	if err != nil {
		return hperrors.Wrap(err).WithMsgLog("failed to add members to zset")
	}
	return nil
}

func ZRangeByScore(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	opt *redis.ZRangeBy,
) ([]string, error) {
	members, err := cmder.ZRangeByScore(ctx, key, opt).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err).WithMsgLog("failed to get members by score from zset")
	}
	return members, nil
}

// ZRangeByScoreWithScores is ZRangeByScore with each member's score.
func ZRangeByScoreWithScores(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	opt *redis.ZRangeBy,
) ([]redis.Z, error) {
	members, err := cmder.ZRangeByScoreWithScores(ctx, key, opt).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err).WithMsgLog("failed to get members by score from zset")
	}
	return members, nil
}

func ZRem(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	members ...any,
) error {
	if len(members) == 0 {
		return nil
	}
	_, err := cmder.ZRem(ctx, key, members...).Result()
	if err != nil {
		return hperrors.Wrap(err).WithMsgLog("failed to remove members from zset")
	}
	return nil
}

func ZCard(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
) (int64, error) {
	count, err := cmder.ZCard(ctx, key).Result()
	if err != nil {
		return 0, hperrors.Wrap(err).WithMsgLog("failed to get card of zset")
	}
	return count, nil
}
