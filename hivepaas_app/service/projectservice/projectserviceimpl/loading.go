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
			return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", id)
		}
	}

	return projects, nil
}

func (s *service) LoadProjectEnv(
	ctx context.Context,
	db database.IDB,
	projectID, projectEnvID string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption, // NOTE: make sure to add SelectRelation("Project")
) (*entity.ProjectEnv, error) {
	env, err := s.projectEnvRepo.GetByID(ctx, db, projectID, projectEnvID, extraOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if err = s.validateProjectEnvStatus(env, requireProjectActive, requireAppActive); err != nil {
		return nil, apperrors.Wrap(err)
	}
	return env, nil
}

func (s *service) validateProjectEnvStatus(
	env *entity.ProjectEnv,
	requireProjectActive, requireEnvActive bool,
) error {
	projectName := env.ProjectID
	if env.Project != nil {
		projectName = env.Project.Name
	}
	if requireProjectActive && (env.Project == nil || env.Project.Status != base.ProjectStatusActive) {
		return apperrors.Wrap(apperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireEnvActive {
		if env.Status != base.ProjectStatusActive {
			return apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithNTParam("Project", projectName).
				WithNTParam("Env", env.Name)
		}
	}
	return nil
}
