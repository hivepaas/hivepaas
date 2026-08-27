package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadProjectEnv(
	ctx context.Context,
	db database.IDB,
	projectID, projectEnvID string,
	requireProjectActive, requireAppActive bool,
	extraOpts ...bunex.SelectQueryOption, // NOTE: make sure to add SelectRelation("Project")
) (*entity.ProjectEnv, error) {
	env, err := s.projectEnvRepo.GetByID(ctx, db, projectID, projectEnvID, extraOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.validateProjectEnvStatus(env, requireProjectActive, requireAppActive); err != nil {
		return nil, hperrors.Wrap(err)
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
		return hperrors.Wrap(hperrors.ErrProjectInactive).WithNTParam("Name", projectName)
	}
	if requireEnvActive {
		if env.Status != base.ProjectStatusActive {
			return hperrors.Wrap(hperrors.ErrProjectEnvInactive).WithNTParam("Project", projectName).
				WithNTParam("Env", env.Name)
		}
	}
	return nil
}
