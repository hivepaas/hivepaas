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
		bunex.SelectRelation("ProjectEnv"),
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
			bunex.SelectWhereIn("app.status IN (?)", req.Status...),
		)
	}
	if len(req.Env) > 0 {
		listOpts = append(listOpts,
			bunex.SelectJoin("JOIN project_envs ON project_envs.id = app.project_env_id"),
			bunex.SelectWhereIn("project_envs.name IN (?)", req.Env...),
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
			bunex.SelectWhereIn("app.id IN (?)", auth.AllowObjectIDs...),
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
