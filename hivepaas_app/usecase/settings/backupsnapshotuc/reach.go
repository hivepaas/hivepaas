package backupsnapshotuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

// snapshotReach is what a view may list: the snapshots of repoIDs owned by an
// app of appIDs, or, owned by no app that still exists, those kept in a
// repository of ownerRepoIDs.
type snapshotReach struct {
	repoIDs      []string
	ownerRepoIDs []string
	appIDs       []string
}

// computeReach decides a view's reach from the repositories and apps it takes
// in, and whether its viewer may read an env (projectID, env).
//
// Every snapshot has an owner: the app its hivepaas.app tag names, or its
// repository when it has none or that app is gone. A view keeps the snapshots
// whose owner is inside it and readable by its viewer, so a project's view never
// shows another project's app through a global repository they share.
func computeReach(
	scope *entity.ObjectScope,
	repos []*entity.Setting,
	apps []*entity.App,
	allows func(projectID, env string) bool,
) *snapshotReach {
	reach := &snapshotReach{}
	for _, repo := range repos {
		reach.repoIDs = append(reach.repoIDs, repo.ID)
		if ownsRepo(scope, repo, allows) {
			reach.ownerRepoIDs = append(reach.ownerRepoIDs, repo.ID)
		}
	}
	for _, app := range apps {
		projectID, env := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
		if projectID == "" {
			projectID = app.ProjectID
		}
		if allows(projectID, env) {
			reach.appIDs = append(reach.appIDs, app.ID)
		}
	}
	return reach
}

// ownsRepo says whether a view holds a repository's own snapshots - those no
// app owns - and its viewer may read them.
func ownsRepo(scope *entity.ObjectScope, repo *entity.Setting, allows func(projectID, env string) bool) bool {
	switch scope.ScopeType { //nolint:exhaustive
	case base.ObjectScopeGlobal:
		return repo.Scope == base.ObjectScopeGlobal
	case base.ObjectScopeProject:
		if repo.Scope == base.ObjectScopeProject {
			return repo.ObjectID == scope.ProjectID
		}
		if repo.Scope == base.ObjectScopeProjectEnv {
			projectID, env := projecthelper.ParseProjectEnvID(repo.ObjectID)
			return projectID == scope.ProjectID && allows(projectID, env)
		}
		return false
	case base.ObjectScopeProjectEnv:
		return repo.Scope == base.ObjectScopeProjectEnv && repo.ObjectID == scope.ProjectEnvID
	}
	// An app's view shows its own snapshots only.
	return false
}
