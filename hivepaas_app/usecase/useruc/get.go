package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		loadOpts = append(loadOpts,
			bunex.SelectRelation("Accesses.ResourceProject"),
		)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID, loadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := userdto.TransformUserDetails(user)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.GetUserResp{
		Data: resp,
	}, nil
}
