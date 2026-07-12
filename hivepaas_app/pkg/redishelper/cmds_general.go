package redishelper

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func ScanKeys(
	ctx context.Context,
	cmder Cmdable,
	pattern string,
) (allKeys []string, err error) {
	var cursor uint64
	for {
		var keys []string
		keys, cursor, err = cmder.Scan(ctx, cursor, pattern, 100).Result() //nolint:mnd
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	return allKeys, nil
}

func Get[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
) (value T, err error) {
	valueStr, err := cmder.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return value, apperrors.NewNotFoundNT(key)
		}
		return value, apperrors.NewNotFoundNT(key)
	}
	return unmarshalStr[T](valueStr)
}

func MGet[T any](
	ctx context.Context,
	cmder Cmdable,
	keys ...string,
) (values []T, err error) {
	if len(keys) == 0 {
		return values, nil
	}
	slice, err := cmder.MGet(ctx, keys...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err)
	}
	return unmarshalSlice[T](slice...)
}

func Set[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	value T,
	expiration time.Duration,
) (err error) {
	data, err := jsonMarshal(value)
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to marshal value")
	}
	_, err = cmder.Set(ctx, key, data, expiration).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to set value in redis")
	}
	return nil
}

func SetXX[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	value T,
	expiration time.Duration,
) (err error) {
	data, err := jsonMarshal(value)
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to marshal value")
	}
	_, err = cmder.SetXX(ctx, key, data, expiration).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to set value in redis")
	}
	return nil
}

func SetNX[T any](
	ctx context.Context,
	cmder Cmdable,
	key string,
	value T,
	expiration time.Duration,
) (err error) {
	data, err := jsonMarshal(value)
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to marshal value")
	}
	_, err = cmder.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to set value in redis")
	}
	return nil
}

func MSet[T any](
	ctx context.Context,
	cmder Cmdable,
	keys []string,
	values []T,
	expiration time.Duration,
) (err error) {
	if len(keys) == 0 {
		return nil
	}
	_, err = cmder.TxPipelined(ctx, func(p redis.Pipeliner) error {
		for i, key := range keys {
			data, err := jsonMarshal(values[i])
			if err != nil {
				return apperrors.Wrap(err).WithMsgLog("failed to marshal value")
			}
			if expiration == redis.KeepTTL {
				_, err = p.SetXX(ctx, key, data, expiration).Result()
			} else {
				_, err = p.Set(ctx, key, data, expiration).Result()
			}
			if err != nil {
				return apperrors.Wrap(err).WithMsgLog("failed to set value in redis")
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func Del(
	ctx context.Context,
	cmder Cmdable,
	keys ...string,
) (err error) {
	_, err = cmder.Del(ctx, keys...).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to delete value in redis")
	}
	return nil
}

func Exists(
	ctx context.Context,
	cmder Cmdable,
	key string,
) (bool, error) {
	count, err := cmder.Exists(ctx, key).Result()
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func Expire(
	ctx context.Context,
	cmder Cmdable,
	key string,
	expiration time.Duration,
) (err error) {
	_, err = cmder.Expire(ctx, key, expiration).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to expire key")
	}
	return nil
}

func ExpireXX(
	ctx context.Context,
	cmder Cmdable,
	key string,
	expiration time.Duration,
) (err error) {
	_, err = cmder.ExpireXX(ctx, key, expiration).Result()
	if err != nil {
		return apperrors.Wrap(err).WithMsgLog("failed to expire key")
	}
	return nil
}
