package appcloneserviceimpl

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) applyEnvVars(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	app := data.TargetApp
	envResp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
		App: app,
		BuildOptions: envvarservice.EnvBuildOptions{
			Sort: true,
		},
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	envVars := make([]string, 0, len(envResp.EnvVars))
	var errs []string
	for _, env := range envResp.EnvVars {
		envVars = append(envVars, env.ToString("="))
		for _, e := range env.Errors {
			errs = append(errs, e.ErrorWithApp(app.Name))
		}
	}

	if len(errs) > 0 {
		return apperrors.Wrap(apperrors.ErrEnvVarContainInvalidReference).WithDisplayLevelHigh().
			WithExtraDetail("%s", strings.Join(errs, "\n"))
	}

	service, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service.Spec.TaskTemplate.ContainerSpec.Env = envVars

	_, err = s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
