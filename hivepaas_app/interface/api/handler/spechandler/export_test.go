package spechandler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// Gating every route on the system module would be wrong in both directions at
// once: it locks a project member out of exporting their own project, and lets
// anyone with system read export a project they cannot otherwise see.
func TestAccessCheckForScopeMatchesWhatIsExported(t *testing.T) {
	const (
		projectID = "01JAB9XED0GTXBSQDFVYAJ8WB1"
		envID     = "01JAB9XED0GTXBSQDFVYAJ8WB1:dev"
		appID     = "01JAB9XED0GTXBSQDFVYAJ8WD1"
	)

	t.Run("app", func(t *testing.T) {
		check, ok := accessCheckForScope(&specdto.ExportSpecReq{
			ProjectID: projectID, ProjectEnvID: envID, AppID: appID,
		}).(*permission.AppAccessCheck)
		assert.True(t, ok, "an app export is gated on the app")
		assert.Equal(t, projectID, check.ProjectID)
		assert.Equal(t, envID, check.ProjectEnv)
		assert.Equal(t, appID, check.AppID)
		assert.Equal(t, base.ActionTypeRead, check.Action)
	})

	t.Run("env", func(t *testing.T) {
		check, ok := accessCheckForScope(&specdto.ExportSpecReq{
			ProjectID: projectID, ProjectEnvID: envID,
		}).(*permission.ProjectAccessCheck)
		assert.True(t, ok, "an env export is gated on the project and env")
		assert.Equal(t, projectID, check.ProjectID)
		assert.NotNil(t, check.ProjectEnv)
		assert.Equal(t, envID, *check.ProjectEnv)
	})

	t.Run("project", func(t *testing.T) {
		check, ok := accessCheckForScope(&specdto.ExportSpecReq{
			ProjectID: projectID,
		}).(*permission.ProjectAccessCheck)
		assert.True(t, ok, "a project export is gated on the project")
		assert.Equal(t, projectID, check.ProjectID)
		assert.Nil(t, check.ProjectEnv, "and on no particular env")
	})

	// The global export genuinely crosses every project, which is the same
	// reason the audit log listing uses the system module.
	t.Run("global", func(t *testing.T) {
		check, ok := accessCheckForScope(&specdto.ExportSpecReq{}).(*permission.GeneralResourceAccessCheck)
		assert.True(t, ok)
		assert.Equal(t, base.ResourceModuleSystem, check.Module)
		assert.Equal(t, base.ResourceTypeSetting, check.ResourceType)
	})
}

// A project export must never fall through to the system-module check, which
// would let system read stand in for project access.
func TestAccessCheckForScopeNeverUsesTheSystemModuleForAScopedExport(t *testing.T) {
	for name, req := range map[string]*specdto.ExportSpecReq{
		"project": {ProjectID: "p1"},
		"env":     {ProjectID: "p1", ProjectEnvID: "p1:dev"},
		"app":     {ProjectID: "p1", ProjectEnvID: "p1:dev", AppID: "a1"},
	} {
		_, isGeneral := accessCheckForScope(req).(*permission.GeneralResourceAccessCheck)
		assert.False(t, isGeneral, "%s export must not be gated on the system module", name)
	}
}
