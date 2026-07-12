package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) ListAppBase(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.ListAppBaseReq,
) (*appdto.ListAppBaseResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	}

	if req.ParentID != "" {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.parent_id = ?", req.ParentID),
		)
	} else {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.parent_id IS NULL"),
		)
	}
	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.status IN (?)", bunex.List(req.Status)),
		)
	}
	if len(req.Env) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.env IN (?)", bunex.List(req.Env)),
		)
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("app.name ILIKE ?", keyword),
				bunex.SelectWhereOr("app.note ILIKE ?", keyword),
			),
		)
	}
	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app.id IN (?)", bunex.List(auth.AllowObjectIDs)),
		)
	}

	apps, pagingMeta, err := uc.appRepo.List(ctx, uc.db, req.ProjectID, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.ListAppBaseResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: appdto.TransformAppsBase(apps),
	}, nil
}
