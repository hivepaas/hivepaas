package appdeploymentuc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// renderList builds the SQL the deployments of app APP are listed with, for the
// statuses asked, without a database. D1 is in progress.
func renderList(statuses ...base.DeploymentStatus) string {
	db := bun.NewDB(nil, pgdialect.New())
	var deployments []*entity.Deployment
	query := db.NewSelect().Model(&deployments).Where("deployment.app_id = ?", "APP")
	return bunex.ApplySelect(query, statusFilter(statuses, []string{"D1"})...).String()
}

// In progress and done: the cached in-progress ones, or a done one - of this app
// both. The statuses were once passed as one value, a JSON array no row
// matched, and the alternative reached past the app's own filter.
func TestStatusFilterInProgressOrAnotherStaysWithinTheApp(t *testing.T) {
	sql := renderList(base.DeploymentStatusInProgress, base.DeploymentStatusDone)

	assert.Contains(t, sql, `deployment.status IN ('done')`)
	assert.NotContains(t, sql, `'["`)
	assert.Contains(t, sql, `(deployment.app_id = 'APP') AND ((deployment.id IN ('D1')) OR `+
		`((deployment.id NOT IN ('D1')) AND (deployment.status IN ('done'))))`)
}

func TestStatusFilterWithoutInProgressReadsTheRows(t *testing.T) {
	sql := renderList(base.DeploymentStatusDone, base.DeploymentStatusFailed)

	assert.Contains(t, sql, `(deployment.app_id = 'APP') AND ((deployment.id NOT IN ('D1')) AND `+
		`(deployment.status IN ('done', 'failed')))`)
}

func TestStatusFilterInProgressAloneReadsTheCache(t *testing.T) {
	sql := renderList(base.DeploymentStatusInProgress)

	assert.Contains(t, sql, `(deployment.app_id = 'APP') AND (deployment.id IN ('D1'))`)
	assert.NotContains(t, sql, `deployment.status`)
}

func TestStatusFilterNothingAskedFiltersNothing(t *testing.T) {
	assert.NotContains(t, renderList(), `deployment.status`)
	assert.NotContains(t, renderList(), `deployment.id`)
}
