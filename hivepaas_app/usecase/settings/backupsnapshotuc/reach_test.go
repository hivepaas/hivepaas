package backupsnapshotuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func repoAt(id string, scope base.ObjectScopeType, objectID string) *entity.Setting {
	return &entity.Setting{ID: id, Type: base.SettingTypeBackupRepo, Scope: scope, ObjectID: objectID}
}

func appIn(id, projectID, env string) *entity.App {
	return &entity.App{ID: id, ProjectID: projectID, ProjectEnvID: projectID + ":" + env}
}

var (
	globalRepo  = repoAt("r-global", base.ObjectScopeGlobal, "")
	projectRepo = repoAt("r-p1", base.ObjectScopeProject, "p1")
	devRepo     = repoAt("r-p1-dev", base.ObjectScopeProjectEnv, "p1:dev")
	prodRepo    = repoAt("r-p1-prod", base.ObjectScopeProjectEnv, "p1:prod")

	webDev  = appIn("web-dev", "p1", "dev")
	webProd = appIn("web-prod", "p1", "prod")
	other   = appIn("other", "p2", "dev")
)

func allowAll(string, string) bool { return true }

// The global view reaches every global repository's snapshots: owned by any app
// the viewer may read, or by the repository.
func TestGlobalReach(t *testing.T) {
	reach := computeReach(&entity.ObjectScope{ScopeType: base.ObjectScopeGlobal},
		[]*entity.Setting{globalRepo}, []*entity.App{webDev, webProd, other},
		func(projectID, _ string) bool { return projectID == "p1" })

	assert.Equal(t, []string{"r-global"}, reach.repoIDs)
	assert.Equal(t, []string{"r-global"}, reach.ownerRepoIDs)
	assert.Equal(t, []string{"web-dev", "web-prod"}, reach.appIDs, "not the apps of a project the viewer cannot read")
}

// A project's view: its apps, and the repositories of the project and of the
// envs its viewer may read - never another project's apps, nor a global
// repository's own snapshots.
func TestProjectReach(t *testing.T) {
	reach := computeReach(&entity.ObjectScope{ScopeType: base.ObjectScopeProject, ProjectID: "p1"},
		[]*entity.Setting{globalRepo, projectRepo, devRepo, prodRepo}, []*entity.App{webDev, webProd},
		func(_, env string) bool { return env == "dev" })

	assert.ElementsMatch(t, []string{"r-global", "r-p1", "r-p1-dev", "r-p1-prod"}, reach.repoIDs)
	assert.ElementsMatch(t, []string{"r-p1", "r-p1-dev"}, reach.ownerRepoIDs)
	assert.Equal(t, []string{"web-dev"}, reach.appIDs)
}

// An env's view: its apps, and its own repositories' snapshots.
func TestEnvReach(t *testing.T) {
	reach := computeReach(&entity.ObjectScope{ScopeType: base.ObjectScopeProjectEnv, ProjectID: "p1",
		ProjectEnvID: "p1:dev"}, []*entity.Setting{globalRepo, projectRepo, devRepo}, []*entity.App{webDev}, allowAll)

	assert.Equal(t, []string{"r-p1-dev"}, reach.ownerRepoIDs)
	assert.Equal(t, []string{"web-dev"}, reach.appIDs)
}

// An app's view: its own snapshots, from any repository it sees.
func TestAppReach(t *testing.T) {
	reach := computeReach(&entity.ObjectScope{ScopeType: base.ObjectScopeApp, ProjectID: "p1",
		ProjectEnvID: "p1:dev", AppID: "web-dev"}, []*entity.Setting{globalRepo, projectRepo, devRepo},
		[]*entity.App{webDev}, allowAll)

	assert.Len(t, reach.repoIDs, 3)
	assert.Empty(t, reach.ownerRepoIDs)
	assert.Equal(t, []string{"web-dev"}, reach.appIDs)
}
