package sysstatusuc

import (
	"context"
	"database/sql"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// The point of these stats is to answer a sizing question from real numbers, so they are worth
// checking against a real pool rather than a stub.
func newTestDB(t *testing.T, maxOpen int) *database.DB {
	t.Helper()
	dsn := os.Getenv("HIVEPAAS_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("HIVEPAAS_TEST_PG_DSN not set, skipping DB stats test")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Skipf("bad dsn: %v", err)
	}
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port))), 2*time.Second)
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	_ = conn.Close()

	sqlDB := sql.OpenDB(stdlib.GetConnector(*cfg))
	db := bun.NewDB(sqlDB, pgdialect.New())
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)
	return &database.DB{DB: db}
}

func TestGetDBStats_ReportsTheConfiguredCeiling(t *testing.T) {
	db := newTestDB(t, 7)
	defer func() { _ = db.Close() }()

	resp, err := New(db).GetDBStats(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, 7, resp.Data.MaxOpenConnections)
	assert.GreaterOrEqual(t, resp.Data.OpenConnections, 0)
}

// WaitCount is the number the sizing decision hangs on: it has to stay at zero while the pool is
// big enough, and rise once callers actually queue for a connection.
func TestGetDBStats_WaitCountTracksPoolContention(t *testing.T) {
	const poolSize = 2
	db := newTestDB(t, poolSize)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	uc := New(db)

	// A single query cannot contend with itself, so nothing should have waited yet.
	var one int
	assert.NoError(t, db.QueryRowContext(ctx, "SELECT 1").Scan(&one))

	before, err := uc.GetDBStats(ctx, nil)
	assert.NoError(t, err)
	assert.Zero(t, before.Data.WaitCount, "an uncontended pool must report no waiting")

	// Now ask for more concurrent work than the pool can serve at once.
	var wg sync.WaitGroup
	for range poolSize * 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var n int
			_ = db.QueryRowContext(ctx, "SELECT pg_sleep(0.2), 1").Scan(&n, &n)
		}()
	}
	wg.Wait()

	after, err := uc.GetDBStats(ctx, nil)
	assert.NoError(t, err)
	assert.Positive(t, after.Data.WaitCount, "callers queued for a connection, so WaitCount must rise")
	assert.Positive(t, after.Data.WaitDurationMs)
	assert.NotEmpty(t, after.Data.WaitDurationHuman)
}
