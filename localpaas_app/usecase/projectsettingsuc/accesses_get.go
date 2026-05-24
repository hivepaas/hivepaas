package projectsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) GetUserAccesses(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.GetUserAccessesReq,
) (*projectsettingsdto.GetUserAccessesResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	objPerms, modPerms, err := uc.permissionManager.LoadObjectAccesses(ctx, uc.db, &permission.AccessCheck{
		SubjectType:    base.SubjectTypeUser,
		ResourceModule: base.ResourceModuleProject,
		ResourceType:   base.ResourceTypeProject,
		ResourceID:     project.ID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
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
