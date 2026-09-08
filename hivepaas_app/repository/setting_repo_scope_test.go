package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// renderSettingQuery builds the SQL a scope filter produces, without a database.
func renderSettingQuery(opts []bunex.SelectQueryOption) string {
	db := bun.NewDB(nil, pgdialect.New())
	var settings []*entity.Setting
	return bunex.ApplySelect(db.NewSelect().Model(&settings), opts...).String()
}

func appScope() *entity.ObjectScope {
	return &entity.ObjectScope{
		ScopeType: base.ObjectScopeApp, AppID: "APP",
		ProjectID: "PROJ", ProjectEnvID: "ENV", ParentAppID: "PARENT",
	}
}

// Imported settings are matched with a semi-join, never a join.
//
// shared_settings is keyed (object_id, setting_id), so one setting can be shared
// into several of the scopes a single filter looks at - a global secret imported
// into a project and into one of its envs. Joining returns it once per share:
// verified against a real database, the join form returned such a setting twice
// and this form once. The page would carry the duplicate and Count() would agree
// with the page rather than with reality.
func TestSettingScopeFiltersDoNotJoinSharedSettings(t *testing.T) {
	repo := &settingRepo{}
	scope := appScope()

	for name, sql := range map[string]string{
		"app":     renderSettingQuery(repo.applyAppFilter(nil, scope)),
		"env":     renderSettingQuery(repo.applyProjectEnvFilter(nil, scope)),
		"project": renderSettingQuery(repo.applyProjectFilter(nil, scope)),
	} {
		t.Run(name, func(t *testing.T) {
			assert.NotContains(t, sql, "JOIN shared_settings",
				"a join can return one setting once per share - use the semi-join")
			assert.Contains(t, sql, "EXISTS (SELECT 1 FROM shared_settings",
				"imported settings still have to be matched")
		})
	}
}

// A scope with no parent app must not compare against an empty id.
//
// object_id is NOT NULL in shared_settings, so "" can never match there; in
// settings it would match a row whose object_id was written as "" rather than
// NULL, which is a global setting reached without the inheritable check.
func TestSettingScopeFilterDropsAbsentObjectIDs(t *testing.T) {
	repo := &settingRepo{}
	scope := appScope()
	scope.ParentAppID = ""

	sql := renderSettingQuery(repo.applyAppFilter(nil, scope))

	assert.NotContains(t, sql, "''", "an absent parent app must not be compared against")
	assert.Contains(t, sql, "IN ('ENV', 'PROJ')")
}

// The inheritable flag has to gate every inherited source.
//
// It has already escaped once: the global branch sat outside the AND, so any
// global setting was visible in every scope whether it was marked inheritable or
// not. AND binds tighter than OR, so that mistake is invisible in the Go and
// obvious in the SQL - which is why this reads the SQL.
func TestInheritedSettingsRequireInheritable(t *testing.T) {
	sql := renderSettingQuery((&settingRepo{}).applyAppFilter(nil, appScope()))

	assert.Contains(t, sql, "(setting.inheritable = TRUE) AND (",
		"every inherited source must sit inside the AND, global included")

	// The global branch is the one that escaped, so pin its position: it may only
	// appear after the flag, which is what puts it inside the AND.
	flagAt := strings.Index(sql, "setting.inheritable = TRUE")
	globalAt := strings.Index(sql, "setting.object_id IS NULL")
	assert.Greater(t, globalAt, flagAt,
		"global settings must not be reachable before the inheritable flag is checked")
}
