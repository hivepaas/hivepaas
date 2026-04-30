package transaction

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tracerr"
)

type options struct {
	isolation      sql.IsolationLevel
	retryTimes     uint
	maxRetryTimes  uint
	retryDelay     time.Duration
	checkRetryable func(error) bool
}

func defaultOptions() *options {
	return &options{
		isolation:      sql.LevelDefault,
		retryTimes:     0,
		maxRetryTimes:  3, //nolint:mnd
		retryDelay:     0,
		checkRetryable: IsErrorDeadLock,
	}
}

func MaxRetryTimes(times uint) func(*options) {
	return func(r *options) {
		r.maxRetryTimes = times
	}
}

func NoRetry() func(*options) {
	return func(r *options) {
		r.maxRetryTimes = 0
	}
}

func RetryDelay(delay time.Duration) func(*options) {
	return func(r *options) {
		r.retryDelay = delay
	}
}

func IsolationLevel(level sql.IsolationLevel) func(*options) {
	return func(r *options) {
		r.isolation = level
	}
}

// Execute executes function within a transaction.
// In case error occurs, depending on the check if error should be retried, transaction can be
// executed again. Usually we can retry the transaction when `deadlock` detected as that kind of
// error is not logical one, instead it's a technical limitation of DB.
func Execute(ctx context.Context, db database.IDB, exec func(tx database.Tx) error, ops ...func(*options)) error {
	opts := defaultOptions()
	for _, op := range ops {
		op(opts)
	}

	for {
		err := db.RunInTx(ctx, &sql.TxOptions{
			Isolation: opts.isolation,
		}, func(ctx context.Context, tx bun.Tx) error {
			return exec(database.Tx{Tx: &tx})
		})
		if err == nil {
			return nil
		}

		if opts.retryTimes < opts.maxRetryTimes && opts.checkRetryable(err) {
			if opts.retryDelay > 0 {
				time.Sleep(opts.retryDelay)
			}
			opts.retryTimes++
			continue
		}

		return tracerr.Wrap(err)
	}
}

// IsErrorDeadLock Postgres deadlock check
func IsErrorDeadLock(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == "40P01"
}
