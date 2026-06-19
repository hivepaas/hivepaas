package transaction

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

// mockIDB implements database.IDB by embedding it.
// This satisfies the interface at compile time.
// Since Execute only calls RunInTx, we only need to implement that.
type mockIDB struct {
	database.IDB
	runInTx func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error
}

func (m *mockIDB) RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
	return m.runInTx(ctx, opts, fn)
}

func TestExecute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		db := &mockIDB{
			runInTx: func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
				return fn(ctx, bun.Tx{})
			},
		}

		called := false
		err := Execute(context.Background(), db, func(tx database.Tx) error {
			called = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("retry on deadlock and succeed", func(t *testing.T) {
		calls := 0
		deadlockErr := &pgconn.PgError{Code: pgErrCodeDeadlock}

		db := &mockIDB{
			runInTx: func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
				calls++
				if calls == 1 {
					return deadlockErr
				}
				return fn(ctx, bun.Tx{})
			},
		}

		err := Execute(context.Background(), db, func(tx database.Tx) error {
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("fail after max retries", func(t *testing.T) {
		calls := 0
		deadlockErr := &pgconn.PgError{Code: pgErrCodeDeadlock}

		db := &mockIDB{
			runInTx: func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
				calls++
				return deadlockErr
			},
		}

		err := Execute(context.Background(), db, func(tx database.Tx) error {
			return nil
		}, MaxRetryTimes(2))

		assert.Error(t, err)
		assert.Equal(t, 3, calls) // 1 initial + 2 retries
	})

	t.Run("non-retryable error", func(t *testing.T) {
		otherErr := errors.New("other error")

		db := &mockIDB{
			runInTx: func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
				return otherErr
			},
		}

		err := Execute(context.Background(), db, func(tx database.Tx) error {
			return nil
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, otherErr))
	})

	t.Run("retry delay", func(t *testing.T) {
		calls := 0
		deadlockErr := &pgconn.PgError{Code: pgErrCodeDeadlock}
		delay := 10 * time.Millisecond

		db := &mockIDB{
			runInTx: func(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
				calls++
				if calls == 1 {
					return deadlockErr
				}
				return fn(ctx, bun.Tx{})
			},
		}

		start := time.Now()
		err := Execute(context.Background(), db, func(tx database.Tx) error {
			return nil
		}, RetryDelay(delay))

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, time.Since(start), delay)
	})
}

func TestIsErrorDeadLock(t *testing.T) {
	assert.True(t, IsErrorDeadLock(&pgconn.PgError{Code: pgErrCodeDeadlock}))
	assert.False(t, IsErrorDeadLock(&pgconn.PgError{Code: "XXXXX"}))
	assert.False(t, IsErrorDeadLock(errors.New("generic error")))
}

func TestOptions(t *testing.T) {
	opts := defaultOptions()

	MaxRetryTimes(10)(opts)
	assert.Equal(t, uint(10), opts.maxRetryTimes)

	NoRetry()(opts)
	assert.Equal(t, uint(0), opts.maxRetryTimes)

	RetryDelay(time.Second)(opts)
	assert.Equal(t, time.Second, opts.retryDelay)

	IsolationLevel(sql.LevelSerializable)(opts)
	assert.Equal(t, sql.LevelSerializable, opts.isolation)
}
