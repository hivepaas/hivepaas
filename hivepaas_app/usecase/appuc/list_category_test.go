package appuc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func renderAppQuery(opts ...bunex.SelectQueryOption) string {
	db := bun.NewDB(nil, pgdialect.New())
	var apps []*entity.App
	return bunex.ApplySelect(db.NewSelect().Model(&apps), opts...).String()
}

// The functions of a list are the apps whose kind says so.
func TestAListOfFunctionsIsTheAppsWhoseKindSaysSo(t *testing.T) {
	sql := renderAppQuery(appCategoryFilter([]base.AppCategory{base.AppCategoryFunction}))

	assert.Contains(t, sql, "EXISTS (SELECT 1 FROM settings AS kind WHERE kind.object_id = app.id "+
		"AND kind.type = 'app-kind' AND kind.deleted_at IS NULL AND kind.data->>'category' IN ('function'))")
	assert.NotContains(t, sql, "NOT EXISTS")
}

// An app that declares no kind is a web app, as the kind settings answer it.
func TestAnAppWithoutAKindIsAWebApp(t *testing.T) {
	sql := renderAppQuery(appCategoryFilter([]base.AppCategory{base.AppCategoryWebapp, base.AppCategoryDatabase}))

	assert.Contains(t, sql, "kind.data->>'category' IN ('webapp', 'database'))")
	assert.Contains(t, sql, "OR (NOT EXISTS (SELECT 1 FROM settings AS kind WHERE kind.object_id = app.id "+
		"AND kind.type = 'app-kind' AND kind.deleted_at IS NULL))")
}
