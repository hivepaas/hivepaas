package projectuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
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
			bunex.SelectWhere("project.status IN (?)", bunex.In(req.Status)),
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
			bunex.SelectWhere("project.id IN (?)", bunex.In(auth.AllowObjectIDs)),
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
