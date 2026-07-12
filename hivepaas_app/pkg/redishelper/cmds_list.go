package redishelper

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func RPush[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	values ...T,
) error {
	if len(values) == 0 {
		return nil
	}
	data, err := marshalSlice(values)
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to marshal value")
	}
	_, err = cmder.RPush(ctx, key, data...).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to push values to a list")
	}
	return nil
}

func LRange[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	start, stop int64,
) (values []T, err error) {
	data, err := cmder.LRange(ctx, key, start, stop).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err)
	}
	return unmarshalStrSlice[T](data...)
}

func BLPop(
	ctx context.Context,
	cmder Cmdable,
	keys []string,
	timeout time.Duration,
) (values map[string]string, err error) {
	strSlice, err := cmder.BLPop(ctx, timeout, keys...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err)
	}
	result := make(map[string]string, len(strSlice)/2) //nolint:mnd
	i := 0
	for i < len(strSlice) {
		result[strSlice[i]] = strSlice[i+1]
		i += 2
	}
	return result, nil
}

func BLPopOne[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	timeout time.Duration,
) (val T, err error) {
	strSlice, err := cmder.BLPop(ctx, timeout, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return val, apperrors.NewNotFoundNT(key)
		}
		return val, apperrors.Wrap(err)
	}
	return unmarshalStr[T](strSlice[1])
}
