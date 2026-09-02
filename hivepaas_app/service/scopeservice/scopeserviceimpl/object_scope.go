package scopeserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadObjectScopeData(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
) (err error) {
	requireActive := !scope.NotRequireActive
	switch scope.ScopeType {
	case base.ObjectScopeProject:
		scope.Project, err = s.projectService.LoadProject(ctx, db, scope.ProjectID, requireActive,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectWhereIf(scope.LockScopeObject, "UPDATE OF project"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.ObjectScopeProjectEnv:
		scope.ProjectEnv, err = s.projectService.LoadProjectEnv(ctx, db, scope.ProjectID,
			scope.ProjectEnvID, requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectWhereIf(scope.LockScopeObject, "UPDATE OF project_env"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		scope.Project = scope.ProjectEnv.Project

	case base.ObjectScopeApp:
		scope.App, err = s.appService.LoadApp(ctx, db, scope.ProjectID, scope.AppID,
			requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ProjectEnv"),
			bunex.SelectRelation("ParentApp",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhereIf(scope.LockScopeObject, "UPDATE OF app"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		scope.Project = scope.App.Project
		scope.ProjectEnv = scope.App.ProjectEnv
		scope.ProjectID = scope.App.ProjectID
		scope.ProjectEnvID = scope.App.ProjectEnvID
		scope.ParentAppID = scope.App.ParentID

	case base.ObjectScopeUser:
		scope.User, err = s.userService.LoadUser(ctx, db, scope.UserID, requireActive,
			bunex.SelectWhereIf(scope.LockScopeObject, "UPDATE OF user"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return nil
	}
	return nil
}

func (s *service) LoadObjectScope(
	ctx context.Context,
	db database.IDB,
	scopeType base.ObjectScopeType,
	objectID string,
	requireActive bool,
) (*entity.ObjectScope, error) {
	var scope *entity.ObjectScope
	switch scopeType {
	case base.ObjectScopeProject:
		scope = entity.NewObjectScopeProject(objectID)
	case base.ObjectScopeProjectEnv:
		scope = entity.NewObjectScopeProjectEnv("", objectID)
	case base.ObjectScopeApp:
		scope = entity.NewObjectScopeApp(objectID, "", "", "")
	case base.ObjectScopeUser:
		scope = entity.NewObjectScopeUser(objectID)
	case base.ObjectScopeGlobal:
		scope = entity.NewObjectScopeGlobal()
	case base.ObjectScopeHivepaas:
		scope = entity.NewObjectScopeHivepaas()
	}
	// An unrecognized scope type leaves scope nil, and every line below would panic on it. That is
	// reachable whenever the type comes from stored data rather than a validated request.
	if scope == nil {
		return nil, hperrors.NewUnsupportedNT(scopeType)
	}
	scope.NotRequireActive = !requireActive

	err := s.LoadObjectScopeData(ctx, db, scope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return scope, nil
}
