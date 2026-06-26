package appcopyserviceimpl

import (
	"context"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) applyEnvVars(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	app := data.TargetApp
	envs, _, err := s.envVarService.BuildAppEnvVars(ctx, db, app, false)
	if err != nil {
		return apperrors.New(err)
	}

	envVars := make([]string, 0, len(envs))
	var errs []string
	for _, env := range envs {
		envVars = append(envVars, env.ToString("="))
		errs = append(errs, env.Errors...)
	}

	if len(errs) > 0 {
		return apperrors.New(apperrors.ErrEnvVarContainInvalidReference).WithDisplayLevelHigh().
			WithExtraDetail("%s", strings.Join(errs, "\n"))
	}

	service, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.New(err)
	}
	service.Spec.TaskTemplate.ContainerSpec.Env = envVars

	_, err = s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
