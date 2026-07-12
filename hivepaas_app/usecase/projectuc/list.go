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

func (uc *UC) ListProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.ListProjectReq,
) (*projectdto.ListProjectResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	}
	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("project.status IN (?)", bunex.List(req.Status)),
		)
	}
	// Filter by search keyword
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

	projects, paging, err := uc.projectRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := projectdto.TransformProjects(projects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.ListProjectResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}
