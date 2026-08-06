package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadScopeObject(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
) (err error) {
	requireActive := !scope.NotRequireActive
	switch scope.ScopeType {
	case base.ObjectScopeGlobal:
		return nil

	case base.ObjectScopeProject:
		scope.Project, err = s.projectService.LoadProject(ctx, db, scope.ProjectID, requireActive,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

	case base.ObjectScopeProjectEnv:
		scope.ProjectEnv, err = s.projectService.LoadProjectEnv(ctx, db, scope.ProjectID,
			scope.ProjectEnvID, requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
		)
		if err != nil {
			return apperrors.Wrap(err)
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
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		scope.Project = scope.App.Project
		scope.ProjectEnv = scope.App.ProjectEnv
		scope.ProjectID = scope.App.ProjectID
		scope.ProjectEnvID = scope.App.ProjectEnvID
		scope.ParentAppID = scope.App.ParentID

	case base.ObjectScopeUser:
		scope.User, err = s.userService.LoadUser(ctx, db, scope.UserID, requireActive)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
