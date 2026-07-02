package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) GetProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.GetProjectReq,
) (*projectdto.GetProjectResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("Tags",
			bunex.SelectOrder("display_order"),
		),
		bunex.SelectRelation("Owner",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			// NOTE: for now, we only need to load Envs settings
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeProjectEnvs),
		),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// Loads all accesses of the project
	if req.GetUserAccesses {
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
		project.Accesses = uc.permissionManager.MergeObjectAccessesBySubjectID(objPerms, modPerms)
	}

	resp, err := projectdto.TransformProject(project)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &projectdto.GetProjectResp{
		Data: resp,
	}, nil
}
