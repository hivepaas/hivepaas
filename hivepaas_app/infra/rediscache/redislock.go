package rediscache

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

var (
	LockExpiry         = redsync.WithExpiry
	LockTries          = redsync.WithTries
	LockRetryDelay     = redsync.WithRetryDelay
	LockRetryDelayFunc = redsync.WithRetryDelayFunc
)

type Lock interface {
	Do(ctx context.Context, name string, expiry time.Duration, tries int, retryDelay time.Duration,
		fn func(), options ...redsync.Option) (lockErr error)
}

type lock struct {
	redsync *redsync.Redsync
}

func (rl *lock) Do(
	ctx context.Context,
	name string,
	expiry time.Duration,
	tries int,
	retryDelay time.Duration,
	fn func(),
	options ...redsync.Option,
) (lockErr error) {
	var opts []redsync.Option
	if expiry > 0 {
		opts = append(opts, redsync.WithExpiry(expiry))
	}
	if tries > 0 {
		opts = append(opts, redsync.WithTries(tries))
	}
	if retryDelay > 0 {
		opts = append(opts, redsync.WithRetryDelay(retryDelay))
	}
	opts = append(opts, options...)

	mutex := rl.redsync.NewMutex(name, opts...)
	if err := mutex.LockContext(ctx); err != nil {
		return apperrors.New(err)
	}
	defer mutex.UnlockContext(ctx) //nolint:errcheck

	fn()
	return nil
}

func NewLock(c Client) Lock {
	pool := goredis.NewPool(c)
	return &lock{redsync: redsync.New(pool)}
}
