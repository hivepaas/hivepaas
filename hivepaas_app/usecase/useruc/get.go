package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) GetUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.GetUserReq,
) (*userdto.GetUserResp, error) {
	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}
	if req.GetAccesses {
		// The response reports access per project env, so every env of the projects
		// involved must be loaded, whether the user was granted the project itself
		// or only some of its envs.
		loadOpts = append(loadOpts,
			bunex.SelectRelation("Accesses.ResourceProject.ProjectEnvs"),
			bunex.SelectRelation("Accesses.ResourceProjectEnv.Project.ProjectEnvs"),
		)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID, loadOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := userdto.TransformUserDetails(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.GetUserResp{
		Data: resp,
	}, nil
}
