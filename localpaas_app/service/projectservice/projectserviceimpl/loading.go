package projectserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
)

func (s *service) LoadProject(
	ctx context.Context,
	db database.IDB,
	projectID string,
	requireActive bool,
	extraLoadOpts ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	var loadOpts []bunex.SelectQueryOption
	if requireActive {
		loadOpts = append(loadOpts,
			bunex.SelectWhere("project.status = ?", base.ProjectStatusActive))
	}
	loadOpts = append(loadOpts, extraLoadOpts...)

	project, err := s.projectRepo.GetByID(ctx, db, projectID, loadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return project, nil
}

func (s *service) LoadProjects(
	ctx context.Context,
	db database.IDB,
	projectIDs []string,
	requireActive bool,
	extraLoadOpts ...bunex.SelectQueryOption,
) ([]*entity.Project, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	var loadOpts []bunex.SelectQueryOption
	if requireActive {
		loadOpts = append(loadOpts,
			bunex.SelectWhere("project.status = ?", base.ProjectStatusActive))
	}
	loadOpts = append(loadOpts, extraLoadOpts...)

	projects, err := s.projectRepo.ListByIDs(ctx, db, projectIDs, loadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	projectMap := entityutil.SliceToIDMap(projects)
	for _, id := range projectIDs {
		if _, exist := projectMap[id]; !exist {
			return nil, apperrors.NewNotFound(apperrors.Fmt("Project '%v'", id))
		}
	}

	return projects, nil
}
