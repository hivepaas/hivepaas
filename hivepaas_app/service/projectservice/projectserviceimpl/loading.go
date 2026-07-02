package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
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
		return nil, apperrors.New(err)
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
		return nil, apperrors.New(err)
	}

	projectMap := entityutil.SliceToIDMap(projects)
	for _, id := range projectIDs {
		if _, exist := projectMap[id]; !exist {
			return nil, apperrors.New(apperrors.ErrProjectNotFound).WithParam("Name", id)
		}
	}

	return projects, nil
}
