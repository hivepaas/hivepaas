package dblock

import (
	"context"
	"database/sql"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// These need a real Postgres: advisory locks are the whole point, and no fake reproduces their
// session semantics. Point HIVEPAAS_TEST_PG_DSN at one to enable them.
func newTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("HIVEPAAS_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("HIVEPAAS_TEST_PG_DSN not set, skipping advisory lock test")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Skipf("bad dsn: %v", err)
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.Host, itoa(cfg.Port)), 2*time.Second)
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	_ = conn.Close()

	sqlDB := sql.OpenDB(stdlib.GetConnector(*cfg))
	return &database.DB{DB: bun.NewDB(sqlDB, pgdialect.New())}
}

func itoa(p uint16) string {
	const digits = "0123456789"
	if p == 0 {
		return "0"
	}
	var buf [5]byte
	i := len(buf)
	for p > 0 {
		i--
		buf[i] = digits[p%10]
		p /= 10
	}
	return string(buf[i:])
}

func TestTryAcquire_SecondCallerIsRefusedNotQueued(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	first, acquired, err := TryAcquire(ctx, db, "test:refused")
	assert.NoError(t, err)
	assert.True(t, acquired)
	defer func() { _ = first.Release(ctx) }()

	// The point of TryAcquire: it comes back immediately instead of waiting for the holder.
	start := time.Now()
	second, acquired, err := TryAcquire(ctx, db, "test:refused")
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.False(t, acquired, "a second caller must not get the lock")
	assert.Nil(t, second)
	assert.Less(t, elapsed, time.Second, "TryAcquire must not block on the holder")
}

// The reason for an advisory lock over a locked row: it outlives the transactions taken while it
// is held, which is what lets the caller commit, run slow work, and commit again.
func TestTryAcquire_SurvivesTransactionBoundaries(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	lock, acquired, err := TryAcquire(ctx, db, "test:across-tx")
	assert.NoError(t, err)
	assert.True(t, acquired)
	defer func() { _ = lock.Release(ctx) }()

	// Commit a transaction in the middle, the way the cleanup flow does.
	tx, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)
	_, err = tx.ExecContext(ctx, "SELECT 1")
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

	_, stillFree, err := TryAcquire(ctx, db, "test:across-tx")
	assert.NoError(t, err)
	assert.False(t, stillFree, "the lock must still be held after an unrelated commit")
}

func TestRelease_MakesTheLockAvailableAgain(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	first, acquired, err := TryAcquire(ctx, db, "test:release")
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.NoError(t, first.Release(ctx))

	second, acquired, err := TryAcquire(ctx, db, "test:release")
	assert.NoError(t, err)
	assert.True(t, acquired, "the lock must be free once released")
	assert.NoError(t, second.Release(ctx))

	// Releasing twice is a no-op, so a deferred release after an explicit one is harmless.
	assert.NoError(t, second.Release(ctx))
}

func TestTryAcquire_DifferentNamesDoNotBlockEachOther(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	a, acquired, err := TryAcquire(ctx, db, "test:repo-a")
	assert.NoError(t, err)
	assert.True(t, acquired)
	defer func() { _ = a.Release(ctx) }()

	b, acquired, err := TryAcquire(ctx, db, "test:repo-b")
	assert.NoError(t, err)
	assert.True(t, acquired, "cleanups of different repositories must run in parallel")
	defer func() { _ = b.Release(ctx) }()
}

func TestKey_IsStableAndNameSpecific(t *testing.T) {
	assert.Equal(t, Key("backup-repo:cleanup:abc"), Key("backup-repo:cleanup:abc"))
	assert.NotEqual(t, Key("backup-repo:cleanup:abc"), Key("backup-repo:cleanup:abd"))
}

// A refused acquire has to hand its connection back. Leaking one per refusal would drain the pool
// after a handful of rejected requests.
func TestTryAcquire_RefusedAcquireDoesNotLeakConnections(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	const poolSize = 3
	db.SetMaxOpenConns(poolSize)

	holder, acquired, err := TryAcquire(ctx, db, "test:leak")
	assert.NoError(t, err)
	assert.True(t, acquired)
	defer func() { _ = holder.Release(ctx) }()

	// Far more refusals than the pool can hold. A leak would block here instead of returning.
	for i := range poolSize * 5 {
		callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, acquired, err := TryAcquire(callCtx, db, "test:leak")
		cancel()
		if !assert.NoError(t, err, "refusal %d failed - the pool is likely exhausted", i) {
			return
		}
		assert.False(t, acquired)
	}

	// The pool must still have room for real work.
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var one int
	assert.NoError(t, db.QueryRowContext(callCtx, "SELECT 1").Scan(&one))
	assert.Equal(t, 1, one)
}
