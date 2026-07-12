package redishelper

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func HGet[T any](
	ctx context.Context,
	cmder redis.Cmdable,
	key, field string,
) (value T, err error) {
	data, err := cmder.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return value, apperrors.NewNotFoundNT(key).WithCause(err)
		}
		return value, apperrors.Wrap(err)
	}
	return unmarshalStr[T](data)
}

func HMGet[T any](
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	fields []string,
) (values []T, err error) {
	if len(fields) == 0 {
		return nil, nil
	}
	data, err := cmder.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return unmarshalSlice[T](data...)
}

func HGetAll[T any](
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
) (map[string]T, error) {
	data, err := cmder.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.NewNotFoundNT(key).WithCause(err)
		}
		return nil, apperrors.Wrap(err)
	}
	return unmarshalStrMap[T](data)
}

func HSet[T any](
	ctx context.Context,
	cmder redis.Cmdable,
	key, field string,
	value T,
	expiration time.Duration,
) error {
	return HMSet(ctx, cmder, key, []string{field}, []T{value}, expiration)
}

func HMSet[T any](
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	fields []string,
	values []T,
	expiration time.Duration,
) error {
	if len(fields) == 0 {
		return nil
	}
	data, err := marshalKVSlices(fields, values)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if _, err := cmder.HSet(ctx, key, data...).Result(); err != nil {
		return apperrors.Wrap(err)
	}
	if expiration > 0 {
		if _, err := cmder.HExpire(ctx, key, expiration, fields...).Result(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func HDel(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	fields ...string,
) error {
	if len(fields) == 0 {
		return nil
	}
	_, err := cmder.HDel(ctx, key, fields...).Result()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func HExpire(
	ctx context.Context,
	cmder redis.Cmdable,
	key string,
	exp time.Duration,
	fields ...string,
) error {
	_, err := cmder.HExpire(ctx, key, exp, fields...).Result()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
