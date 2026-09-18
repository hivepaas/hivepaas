package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// renderDeleteByScope builds the SQL DeleteAllByScope runs, without a database.
func renderDeleteByScope(scope base.ObjectScopeType, objectIDs ...string) string {
	repo := &resLinkRepo{}
	db := bun.NewDB(nil, pgdialect.New())
	return repo.deleteAllByScopeQuery(db, scope, objectIDs).String()
}

// Every resource link HivePaaS writes has a setting as its source, so an app's
// links cannot be found through the app's own id. Deleting an app by that id -
// which is what this replaced - matched nothing at all, and left the app's
// domains and ports recorded as taken forever.
func TestDeleteAllByScopeMatchesTheLinksOfTheScopesSettings(t *testing.T) {
	sql := renderDeleteByScope(base.ObjectScopeApp, "APP1", "APP2")

	assert.Contains(t, sql, `res_link.src_type = 'setting'`)
	assert.Contains(t, sql, `FROM "settings" AS "setting"`)
	assert.Contains(t, sql, `setting.scope = 'app'`)
	assert.Contains(t, sql, `setting.object_id IN ('APP1', 'APP2')`)
	assert.Contains(t, sql, `res_link.src_type = 'app' AND res_link.src_id IN ('APP1', 'APP2')`,
		"links the object is itself the source of are deleted too")
}

// The two sources are an alternative to each other, and the whole of it has to
// stay ANDed to what the soft delete adds. Losing those parentheses would match
// every link in the table.
func TestDeleteAllByScopeKeepsItsConditionsTogether(t *testing.T) {
	sql := renderDeleteByScope(base.ObjectScopeApp, "APP1")

	assert.Contains(t, sql, `WHERE ((res_link.src_type = 'setting'`)
	assert.Contains(t, sql, `OR (res_link.src_type = 'app'`)
	assert.Contains(t, sql, `AND "res_link"."deleted_at" IS NULL`)
	assert.Contains(t, sql, `UPDATE "res_links"`, "a soft delete, like every other delete here")
}

// The settings are read including the deleted ones, so this works whether it
// runs before or after they are deleted - which is not the same at every caller.
func TestDeleteAllByScopeFindsSettingsThatAreAlreadyDeleted(t *testing.T) {
	sql := renderDeleteByScope(base.ObjectScopeApp, "APP1")

	assert.NotContains(t, sql, `"setting"."deleted_at" IS NULL`)
}

func TestDeleteAllByScopeCoversEveryScopeThatOwnsSettings(t *testing.T) {
	for scope, srcType := range map[base.ObjectScopeType]string{
		base.ObjectScopeApp:        "app",
		base.ObjectScopeProject:    "project",
		base.ObjectScopeProjectEnv: "project-env",
		base.ObjectScopeUser:       "user",
	} {
		t.Run(string(scope), func(t *testing.T) {
			sql := renderDeleteByScope(scope, "OBJ")
			assert.Contains(t, sql, `setting.scope = '`+string(scope)+`'`)
			assert.Contains(t, sql, `res_link.src_type = '`+srcType+`'`)
		})
	}
}

// A global setting belongs to no object, so there is no object to be the source
// of a link either, and only the settings branch is written.
func TestDeleteAllByScopeWritesNoObjectBranchForGlobalSettings(t *testing.T) {
	sql := renderDeleteByScope(base.ObjectScopeGlobal, "OBJ")

	assert.Contains(t, sql, `res_link.src_type = 'setting'`)
	assert.NotContains(t, sql, " OR ")
}

// renderDeleteAll builds the SQL DeleteAll runs, without a database.
func renderDeleteAll(opts ...bunex.DeleteQueryOption) string {
	db := bun.NewDB(nil, pgdialect.New())
	return bunex.ApplyDelete(db.NewDelete().Model((*entity.ResLink)(nil)), opts...).String()
}

// The orphan cleanup deletes links whose setting is gone. Its condition lives in
// the cleanup service; what this holds is that DeleteAll carries it through as a
// soft delete rather than a hard one.
func TestDeleteAllSoftDeletesWhatTheConditionsMatch(t *testing.T) {
	sql := renderDeleteAll(
		bunex.DeleteWhere("res_link.src_type = ?", base.ResourceTypeSetting),
		bunex.DeleteWhere("NOT EXISTS (SELECT 1 FROM settings"+
			" WHERE settings.id = res_link.src_id AND settings.deleted_at IS NULL)"),
	)

	assert.Contains(t, sql, `UPDATE "res_links"`)
	assert.Contains(t, sql, `SET "deleted_at"`)
	assert.Contains(t, sql, `res_link.src_type = 'setting'`)
	assert.Contains(t, sql, `NOT EXISTS (SELECT 1 FROM settings`)
	assert.Contains(t, sql, `AND "res_link"."deleted_at" IS NULL`)
}

// Deleting every link is never what a caller means, so it takes a condition.
func TestDeleteAllRefusesToRunWithoutACondition(t *testing.T) {
	repo := &resLinkRepo{}

	err := repo.DeleteAll(context.Background(), nil)

	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid)
}
