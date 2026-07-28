package projectsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) GetUserAccesses(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.GetUserAccessesReq,
) (*projectsettingsdto.GetUserAccessesResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("Owner",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnvs",
			bunex.SelectOrder("index"),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	modPerms, projectPerms, envPerms, err := uc.permissionManager.LoadProjectAccesses(ctx, uc.db, project.ID,
		entityutil.ExtractIDs(project.ProjectEnvs), true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := projectsettingsdto.TransformUserAccesses(&projectsettingsdto.UserAccessesTransformInput{
		Project:            project,
		ModulePermissions:  modPerms,
		ProjectPermissions: projectPerms,
		EnvPermissions:     envPerms,
		CurrentUser:        auth.User.User,
	})

	return &projectsettingsdto.GetUserAccessesResp{
		Data: resp,
	}, nil
}
