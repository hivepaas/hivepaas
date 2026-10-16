package repository

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// PlanUsageRepo counts the kinds a plan limits, installation-wide, and takes the
// locks that keep two creations from both taking the last place under a limit.
type PlanUsageRepo interface {
	// LockLimit takes the plan-limit lock of limit until the transaction db ends.
	// Outside a transaction it would be released as soon as it was taken.
	LockLimit(ctx context.Context, db database.IDB, limit base.LicenseLimit) error

	// CountApps counts the apps a plan's limit of apps counts, or, with
	// functions, its limit of functions: neither counts a preview, an app that
	// belongs to another, or a system app.
	CountApps(ctx context.Context, db database.IDB, functions bool) (int, error)
	// CountPreviews counts the previews that exist, without the databases they
	// cloned for themselves.
	CountPreviews(ctx context.Context, db database.IDB) (int, error)
	// CountUsers counts the users that hold a seat: active, or invited and
	// pending.
	CountUsers(ctx context.Context, db database.IDB) (int, error)
	// CountActiveOAuth counts the global OAuth settings of one of kinds that are
	// active.
	CountActiveOAuth(ctx context.Context, db database.IDB, kinds []base.OAuthKind) (int, error)
}

type planUsageRepo struct {
}

func NewPlanUsageRepo() PlanUsageRepo {
	return &planUsageRepo{}
}

func (repo *planUsageRepo) LockLimit(ctx context.Context, db database.IDB, limit base.LicenseLimit) error {
	if _, err := lockLimitQuery(db, limit).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *planUsageRepo) CountApps(ctx context.Context, db database.IDB, functions bool) (int, error) {
	return scanCount(ctx, countAppsQuery(db, functions))
}

func (repo *planUsageRepo) CountPreviews(ctx context.Context, db database.IDB) (int, error) {
	return scanCount(ctx, countPreviewsQuery(db))
}

func (repo *planUsageRepo) CountUsers(ctx context.Context, db database.IDB) (int, error) {
	return scanCount(ctx, countUsersQuery(db))
}

func (repo *planUsageRepo) CountActiveOAuth(ctx context.Context, db database.IDB,
	kinds []base.OAuthKind) (int, error) {
	if len(kinds) == 0 {
		return 0, nil
	}
	return scanCount(ctx, countActiveOAuthQuery(db, kinds))
}

func scanCount(ctx context.Context, query *bun.SelectQuery) (int, error) {
	var count int
	if err := query.Scan(ctx, &count); err != nil {
		return 0, hperrors.Wrap(err)
	}
	return count, nil
}

// lockLimitQuery takes one transaction-level advisory lock per limit, named
// plan-limit:<limit>. Counts are installation-wide and creations rare, so one
// lock per kind is enough: a creation that takes it counts what the one that
// held it before committed.
func lockLimitQuery(db bun.IDB, limit base.LicenseLimit) *bun.RawQuery {
	return db.NewRaw("SELECT pg_advisory_xact_lock(hashtext(?))", "plan-limit:"+string(limit))
}

// planCountedApps is the apps of the projects people make: not deleted, of a
// project not deleted, and not of the hidden project HivePaaS runs its own apps
// in. bun adds app.deleted_at IS NULL itself; the project is joined as a plain
// table, so its deleted_at is written out.
func planCountedApps(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().Model((*entity.App)(nil)).
		ColumnExpr("count(*)").
		Join("JOIN projects AS project ON project.id = app.project_id").
		Where("project.deleted_at IS NULL").
		Where("project.key <> ?", base.HivepaasProjectKey)
}

// countAppsQuery leaves out previews (parent_id) and the apps that belong to
// another (logical_parent_id): a preview's cloned databases and a template's
// dependencies, which app lists leave out too. A function is an app whose kind
// setting says so; an app with no kind setting is a webapp.
func countAppsQuery(db bun.IDB, functions bool) *bun.SelectQuery {
	const isFunction = "EXISTS (SELECT 1 FROM settings AS kind WHERE kind.object_id = app.id " +
		"AND kind.type = ? AND kind.deleted_at IS NULL AND kind.data->>'category' = ?)"
	query := planCountedApps(db).
		Where("app.parent_id IS NULL").
		Where("app.logical_parent_id IS NULL")
	if functions {
		return query.Where(isFunction, base.SettingTypeAppKind, base.AppCategoryFunction)
	}
	return query.Where("NOT "+isFunction, base.SettingTypeAppKind, base.AppCategoryFunction)
}

// countPreviewsQuery counts the apps with a parent and no logical parent. The
// databases a preview clones have a parent too, the app they were cloned from,
// and their logical parent is the preview: they are part of it, not previews.
func countPreviewsQuery(db bun.IDB) *bun.SelectQuery {
	return planCountedApps(db).
		Where("app.parent_id IS NOT NULL").
		Where("app.logical_parent_id IS NULL")
}

// countUsersQuery counts the pending users with the active ones: an invitation
// holds a seat, or a limit of 2 could be invited past. A disabled user holds
// none. bun adds deleted_at IS NULL.
func countUsersQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().Model((*entity.User)(nil)).
		ColumnExpr("count(*)").
		Where(`"user".status IN (?)`, bun.List([]base.UserStatus{base.UserStatusActive, base.UserStatusPending}))
}

// countActiveOAuthQuery reads the global settings, where sign-in providers are,
// as settingRepo's global filter does. bun adds deleted_at IS NULL.
func countActiveOAuthQuery(db bun.IDB, kinds []base.OAuthKind) *bun.SelectQuery {
	return db.NewSelect().Model((*entity.Setting)(nil)).
		ColumnExpr("count(*)").
		Where("setting.scope = ?", base.ObjectScopeGlobal).
		Where("setting.object_id IS NULL").
		Where("setting.type = ?", base.SettingTypeOAuth).
		Where("setting.status = ?", base.SettingStatusActive).
		Where("setting.kind IN (?)", bun.List(kinds))
}
