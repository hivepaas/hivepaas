package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) ListProjectBase(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.ListProjectBaseReq,
) (*projectdto.ListProjectBaseResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	}

	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("project.status IN (?)", bunex.List(req.Status)),
		)
	}

	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("project.name ILIKE ?", keyword),
				bunex.SelectWhereOr("project.note ILIKE ?", keyword),
			),
		)
	}

	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("project.id IN (?)", bunex.List(auth.AllowObjectIDs)),
		)
	}

	if !config.Current.IsDevEnv() {
		listOpts = append(listOpts,
			bunex.SelectWhere("project.key != ?", base.HivepaasProjectKey),
		)
	}

	projects, pagingMeta, err := uc.projectRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &projectdto.ListProjectBaseResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: projectdto.TransformProjectsBase(projects),
	}, nil
}
