package projectsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
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
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	objPerms, modPerms, err := uc.permissionManager.LoadObjectAccesses(ctx, uc.db, &permission.AccessCheck{
		SubjectType:    base.SubjectTypeUser,
		ResourceModule: base.ResourceModuleProject,
		ResourceType:   base.ResourceTypeProject,
		ResourceID:     project.ID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp := projectsettingsdto.TransformUserAccesses(&projectsettingsdto.UserAccessesTransformInput{
		Project:           project,
		ObjectPermissions: objPerms,
		ModulePermissions: modPerms,
		CurrentUser:       auth.User.User,
	})

	return &projectsettingsdto.GetUserAccessesResp{
		Data: resp,
	}, nil
}
