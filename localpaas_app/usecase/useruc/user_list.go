package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) ListUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.ListUserReq,
) (*userdto.ListUserResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}
	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("\"user\".status IN (?)", bunex.In(req.Status)),
		)
	}
	// Filter by search keyword
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("\"user\".email ILIKE ?", keyword),
				bunex.SelectWhereOr("\"user\".username ILIKE ?", keyword),
				bunex.SelectWhereOr("\"user\".full_name ILIKE ?", keyword),
				bunex.SelectWhereOr("\"user\".position ILIKE ?", keyword),
				bunex.SelectWhereOr("\"user\".status = ?", keyword),
				bunex.SelectWhereOr("\"user\".security_option ILIKE ?", keyword),
			),
		)
	}

	users, paging, err := uc.userRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := userdto.TransformUsers(users)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.ListUserResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}
