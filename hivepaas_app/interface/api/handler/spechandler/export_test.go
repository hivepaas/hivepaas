package spechandler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

const (
	testProjectID = "01JAB9XED0GTXBSQDFVYAJ8WB1"
	testAppID     = "01JAB9XED0GTXBSQDFVYAJ8WD1"
)

func reqFor(scope *entity.ObjectScope) *specdto.ExportSpecReq {
	req := specdto.NewExportSpecReq()
	req.Scope = scope
	return req
}

// Gating every route on the system module would be wrong in both directions at
// once: it locks a project member out of exporting their own project, and lets
// anyone with system read export a project they cannot otherwise see.
func TestAccessCheckForScopeMatchesWhatIsExported(t *testing.T) {
	envScope := entity.NewObjectScopeProjectEnv(testProjectID, "dev")

	t.Run("app", func(t *testing.T) {
		raw, err := accessCheckForScope(reqFor(entity.NewObjectScopeApp(testAppID, "", testProjectID, "dev")))
		assert.NoError(t, err)
		check, ok := raw.(*permission.AppAccessCheck)
		assert.True(t, ok, "an app export is gated on the app")
		assert.Equal(t, testProjectID, check.ProjectID)
		assert.Equal(t, envScope.ProjectEnvID, check.ProjectEnv)
		assert.Equal(t, testAppID, check.AppID)
		assert.Equal(t, base.ActionTypeRead, check.Action)
	})

	t.Run("env", func(t *testing.T) {
		raw, err := accessCheckForScope(reqFor(envScope))
		assert.NoError(t, err)
		check, ok := raw.(*permission.ProjectAccessCheck)
		assert.True(t, ok, "an env export is gated on the project and env")
		assert.Equal(t, testProjectID, check.ProjectID)
		assert.NotNil(t, check.ProjectEnv)
		assert.Equal(t, envScope.ProjectEnvID, *check.ProjectEnv)
	})

	t.Run("project", func(t *testing.T) {
		raw, err := accessCheckForScope(reqFor(entity.NewObjectScopeProject(testProjectID)))
		assert.NoError(t, err)
		check, ok := raw.(*permission.ProjectAccessCheck)
		assert.True(t, ok, "a project export is gated on the project")
		assert.Equal(t, testProjectID, check.ProjectID)
		assert.Nil(t, check.ProjectEnv, "and on no particular env")
	})

	// The global export genuinely crosses every project, which is the same
	// reason the audit log listing uses the system module.
	t.Run("global", func(t *testing.T) {
		raw, err := accessCheckForScope(reqFor(entity.NewObjectScopeGlobal()))
		assert.NoError(t, err)
		check, ok := raw.(*permission.GeneralResourceAccessCheck)
		assert.True(t, ok)
		assert.Equal(t, base.ResourceModuleSystem, check.Module)
		assert.Equal(t, base.ResourceTypeSetting, check.ResourceType)
	})
}

// A scoped export must never fall through to the system-module check, which
// would let system read stand in for project access.
func TestAccessCheckForScopeNeverUsesTheSystemModuleForAScopedExport(t *testing.T) {
	for name, scope := range map[string]*entity.ObjectScope{
		"project": entity.NewObjectScopeProject("p1"),
		"env":     entity.NewObjectScopeProjectEnv("p1", "dev"),
		"app":     entity.NewObjectScopeApp("a1", "", "p1", "dev"),
	} {
		raw, err := accessCheckForScope(reqFor(scope))
		assert.NoError(t, err, name)
		_, isGeneral := raw.(*permission.GeneralResourceAccessCheck)
		assert.False(t, isGeneral, "%s export must not be gated on the system module", name)
	}
}

// Scopes that have no export are refused rather than given a default check.
func TestAccessCheckForScopeRefusesScopesWithNoExport(t *testing.T) {
	for name, scope := range map[string]*entity.ObjectScope{
		"user":     entity.NewObjectScopeUser("u1"),
		"hivepaas": entity.NewObjectScopeHivepaas(),
	} {
		check, err := accessCheckForScope(reqFor(scope))
		assert.Error(t, err, name)
		assert.Nil(t, check, name)
	}
}
