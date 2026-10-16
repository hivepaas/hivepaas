package repository

import (
	"context"
	"database/sql"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func renderDB() *bun.DB {
	return bun.NewDB(nil, pgdialect.New())
}

// appsFrom is the start of every count of apps: the apps people made, not
// deleted, of a project not deleted, and not HivePaaS's own.
const appsFrom = `SELECT count(*) FROM "apps" AS "app" JOIN projects AS project ON project.id = app.project_id ` +
	`WHERE (project.deleted_at IS NULL) AND (project.key <> 'hivepaas') `

const isFunction = `EXISTS (SELECT 1 FROM settings AS kind WHERE kind.object_id = app.id ` +
	`AND kind.type = 'app-kind' AND kind.deleted_at IS NULL AND kind.data->>'category' = 'function')`

// Apps leave out previews, the apps that belong to another (a preview's cloned
// databases, a template's dependencies) and functions. An app with no kind
// setting is a webapp, so it is counted.
func TestCountAppsLeavesOutWhatIsNotAnApp(t *testing.T) {
	got := countAppsQuery(renderDB(), false).String()

	assert.Equal(t, appsFrom+`AND (app.parent_id IS NULL) AND (app.logical_parent_id IS NULL) `+
		`AND (NOT `+isFunction+`) AND "app"."deleted_at" IS NULL`, got)
}

func TestCountFunctionsHasTheSameExclusions(t *testing.T) {
	got := countAppsQuery(renderDB(), true).String()

	assert.Equal(t, appsFrom+`AND (app.parent_id IS NULL) AND (app.logical_parent_id IS NULL) `+
		`AND (`+isFunction+`) AND "app"."deleted_at" IS NULL`, got)
}

// A preview's cloned databases have a parent too - the app they were cloned
// from - and the preview as their logical parent. Counted as previews, one
// preview with a database would count twice.
func TestCountPreviewsLeavesOutTheirDatabases(t *testing.T) {
	got := countPreviewsQuery(renderDB()).String()

	assert.Equal(t, appsFrom+`AND (app.parent_id IS NOT NULL) AND (app.logical_parent_id IS NULL) `+
		`AND "app"."deleted_at" IS NULL`, got)
}

// An invitation holds a seat; a disabled user does not.
func TestCountUsersCountsActiveAndPending(t *testing.T) {
	got := countUsersQuery(renderDB()).String()

	assert.Equal(t, `SELECT count(*) FROM "users" AS "user" WHERE ("user".status IN ('active', 'pending')) `+
		`AND "user"."deleted_at" IS NULL`, got)
}

func TestCountActiveOAuthReadsTheGlobalSettings(t *testing.T) {
	got := countActiveOAuthQuery(renderDB(), []base.OAuthKind{base.OAuthKindGoogle, base.OAuthKindOpenIDConnect}).
		String()

	assert.Equal(t, `SELECT count(*) FROM "settings" AS "setting" WHERE (setting.scope = '') `+
		`AND (setting.object_id IS NULL) AND (setting.type = 'oauth') AND (setting.status = 'active') `+
		`AND (setting.kind IN ('google', 'openid-connect')) AND "setting"."deleted_at" IS NULL`, got)
}

// No kind is no query: IN () is not SQL.
func TestCountActiveOAuthOfNoKindIsZero(t *testing.T) {
	count, err := (&planUsageRepo{}).CountActiveOAuth(context.Background(), nil, nil)

	assert.NoError(t, err)
	assert.Zero(t, count)
}

func TestLockLimitIsATransactionLockPerLimit(t *testing.T) {
	assert.Equal(t, `SELECT pg_advisory_xact_lock(hashtext('plan-limit:apps'))`,
		lockLimitQuery(renderDB(), base.LicenseLimitApps).String())
	assert.Equal(t, `SELECT pg_advisory_xact_lock(hashtext('plan-limit:users'))`,
		lockLimitQuery(renderDB(), base.LicenseLimitUsers).String())
}

// newPlanLockTestDB needs a real Postgres, as pkg/dblock's tests do: no fake
// reproduces an advisory lock. Point HIVEPAAS_TEST_PG_DSN at one to run it.
func newPlanLockTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("HIVEPAAS_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("HIVEPAAS_TEST_PG_DSN not set, skipping advisory lock test")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Skipf("bad dsn: %v", err)
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port))), 2*time.Second)
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	_ = conn.Close()

	db := &database.DB{DB: bun.NewDB(sql.OpenDB(stdlib.GetConnector(*cfg)), pgdialect.New())}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// A second creation of a kind waits for the first to end before it counts, and
// a creation of another kind does not wait at all.
func TestLockLimitSerializesCreationsOfAKind(t *testing.T) {
	db := newPlanLockTestDB(t)
	ctx := context.Background()
	repo := NewPlanUsageRepo()
	const waitedFor = 300 * time.Millisecond

	first, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer func() { _ = first.Rollback() }()
	assert.NoError(t, repo.LockLimit(ctx, database.Tx{Tx: &first}, base.LicenseLimitApps))

	other, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer func() { _ = other.Rollback() }()
	assert.NoError(t, repo.LockLimit(ctx, database.Tx{Tx: &other}, base.LicenseLimitUsers),
		"another kind's lock is not held")

	second, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer func() { _ = second.Rollback() }()
	locked := make(chan error, 1)
	go func() { locked <- repo.LockLimit(ctx, database.Tx{Tx: &second}, base.LicenseLimitApps) }()

	select {
	case err = <-locked:
		t.Fatalf("the second creation took the lock while the first held it (err %v)", err)
	case <-time.After(waitedFor):
	}
	assert.NoError(t, first.Commit())
	select {
	case err = <-locked:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the second creation still waits after the first committed")
	}
}
