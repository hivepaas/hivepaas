package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) GetProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.GetProjectReq,
) (*projectdto.GetProjectResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("ProjectEnvs",
			bunex.SelectOrder("index"),
		),
		bunex.SelectRelation("Tags",
			bunex.SelectOrder("index"),
		),
		bunex.SelectRelation("Owner",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Loads all accesses on the project
	if req.GetUserAccesses {
		project.Accesses, err = uc.permissionManager.LoadProjectAccessUsers(ctx, uc.db, project.ID,
			entityutil.ExtractIDs(project.ProjectEnvs))
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	resp, err := projectdto.TransformProject(project)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectdto.GetProjectResp{
		Data: resp,
	}, nil
}
