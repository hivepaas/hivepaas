package backupsnapshotuc

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// renderSnapshotQuery is the SQL the options make, without a database.
func renderSnapshotQuery(opts []bunex.SelectQueryOption) string {
	var settings []*entity.Setting
	return bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&settings), opts...).String()
}

var reachP1 = &snapshotReach{
	repoIDs:      []string{"r-global", "r-p1"},
	ownerRepoIDs: []string{"r-p1"},
	appIDs:       []string{"web-dev"},
}

// A view lists the snapshots of its repositories owned by its apps, or owned by
// no live app and kept in one of its own repositories; newest first.
func TestSnapshotQueryKeepsTheViewsReach(t *testing.T) {
	sql := renderSnapshotQuery(snapshotQueryOpts(reachP1, &snapshotFilter{}))

	assert.Contains(t, sql, `setting.type = 'backup-snapshot'`)
	assert.Contains(t, sql, `setting.ref_id IN ('r-global', 'r-p1')`)
	assert.Contains(t, sql, `t.tag IN ('hivepaas.app:web-dev')`)
	assert.Contains(t, sql, `setting.ref_id IN ('r-p1') AND NOT EXISTS`)
	assert.Contains(t, sql, `JOIN apps AS a ON t.tag = 'hivepaas.app:' || a.id`)
	assert.Contains(t, sql, `ORDER BY (setting.data->>'time')::timestamptz DESC`)
}

// A view that reaches nothing lists nothing.
func TestSnapshotQueryOfAnEmptyReach(t *testing.T) {
	sql := renderSnapshotQuery(snapshotQueryOpts(&snapshotReach{}, &snapshotFilter{}))
	assert.Contains(t, sql, "1=0")
}

// The filters narrow the view, never widen it: a repository outside the reach
// lists nothing.
func TestSnapshotQueryFilters(t *testing.T) {
	from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	sql := renderSnapshotQuery(snapshotQueryOpts(reachP1, &snapshotFilter{
		RepoIDs: []string{"r-p1", "r-elsewhere"},
		AppIDs:  []string{"web-dev"},
		Tags:    []string{"env:prod", "db:main"},
		From:    from,
		Search:  "nightly",
	}))

	assert.Contains(t, sql, `setting.ref_id IN ('r-p1')`)
	assert.NotContains(t, sql, "r-elsewhere")
	assert.Equal(t, 1, strings.Count(sql, `t.tag = 'env:prod'`))
	assert.Equal(t, 1, strings.Count(sql, `t.tag = 'db:main'`))
	assert.Contains(t, sql, `(setting.data->>'time')::timestamptz >= '2026-09-01 00:00:00+00:00'`)
	assert.Contains(t, sql, `setting.name ILIKE '%nightly%'`)

	outside := renderSnapshotQuery(snapshotQueryOpts(reachP1, &snapshotFilter{RepoIDs: []string{"r-elsewhere"}}))
	assert.Contains(t, outside, "1=0")
}
