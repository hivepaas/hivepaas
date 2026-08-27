package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type baseFileData struct {
	ScopeProject    *entity.Project
	ScopeProjectEnv *entity.ProjectEnv
	ScopeApp        *entity.App
	ScopeUser       *entity.User
}

func (uc *UC) loadScopeData(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	data *baseFileData,
) (err error) {
	requireActive := !scope.NotRequireActive
	switch scope.ScopeType {
	case base.ObjectScopeApp:
		data.ScopeApp, err = uc.appService.LoadApp(ctx, db, scope.ProjectID, scope.AppID,
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
			return hperrors.Wrap(err)
		}
		data.ScopeProject = data.ScopeApp.Project

	case base.ObjectScopeProjectEnv:
		data.ScopeProjectEnv, err = uc.projectService.LoadProjectEnv(ctx, db, scope.ProjectID, scope.ProjectEnvID,
			requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.ScopeProject = data.ScopeProjectEnv.Project

	case base.ObjectScopeProject:
		data.ScopeProject, err = uc.projectService.LoadProject(ctx, db, scope.ProjectID, requireActive,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.ObjectScopeUser:
		data.ScopeUser, err = uc.userService.LoadUser(ctx, db, scope.UserID, requireActive)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return nil
	}

	return nil
}
