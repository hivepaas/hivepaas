package projectservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (s *projectService) LoadProject(
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
