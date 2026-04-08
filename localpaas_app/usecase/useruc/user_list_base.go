package useruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) ListUserBase(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.ListUserBaseReq,
) (*userdto.ListUserBaseResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}

	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("\"user\".status IN (?)", bunex.In(req.Status)),
		)
	}

	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhere("\"user\".email ILIKE ?", keyword),
			bunex.SelectWhereOr("\"user\".username ILIKE ?", keyword),
			bunex.SelectWhereOr("\"user\".full_name ILIKE ?", keyword),
			bunex.SelectWhereOr("\"user\".position ILIKE ?", keyword),
			bunex.SelectWhereOr("\"user\".status = ?", keyword),
			bunex.SelectWhereOr("\"user\".security_option ILIKE ?", keyword),
		)
	}

	users, pagingMeta, err := uc.userRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &userdto.ListUserBaseResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: basedto.TransformUsersBase(users),
	}, nil
}
