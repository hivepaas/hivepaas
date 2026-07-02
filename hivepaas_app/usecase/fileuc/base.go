package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type baseFileData struct {
	ScopeProject *entity.Project
	ScopeApp     *entity.App
	ScopeUser    *entity.User
}

func (uc *UC) loadScopeData(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	data *baseFileData,
) (err error) {
	requireActive := !scope.NotRequireActive
	switch scope.ScopeType() {
	case base.ObjectScopeGlobal:
		return nil

	case base.ObjectScopeProject:
		data.ScopeProject, err = uc.projectService.LoadProject(ctx, db, scope.ProjectID, requireActive,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		)
		if err != nil {
			return apperrors.New(err)
		}

	case base.ObjectScopeApp:
		data.ScopeApp, err = uc.appService.LoadApp(ctx, db, scope.ProjectID, scope.AppID,
			requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ParentApp",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		)
		if err != nil {
			return apperrors.New(err)
		}
		data.ScopeProject = data.ScopeApp.Project

	case base.ObjectScopeUser:
		data.ScopeUser, err = uc.userService.LoadUserEx(ctx, db, scope.UserID, requireActive)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
