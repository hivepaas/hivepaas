package projectuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) GetProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.GetProjectReq,
) (*projectdto.GetProjectResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("Tags", bunex.SelectOrder("display_order")),
		bunex.SelectRelation("Apps", bunex.SelectOrder("name")),
		bunex.SelectRelation("Owner"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Loads all accesses of the project
	accesses, err := uc.permissionManager.LoadObjectAccesses(ctx, uc.db, &permission.AccessCheck{
		SubjectType:    base.SubjectTypeUser,
		ResourceModule: base.ResourceModuleProject,
		ResourceType:   base.ResourceTypeProject,
		ResourceID:     project.ID,
		Action:         base.ActionTypeRead,
	}, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	project.Accesses = accesses

	resp, err := projectdto.TransformProject(project)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.GetProjectResp{
		Data: resp,
	}, nil
}
