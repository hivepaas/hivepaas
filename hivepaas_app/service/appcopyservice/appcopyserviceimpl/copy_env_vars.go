package appcopyserviceimpl

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
	data *appCopyData,
) (err error) {
	app := data.TargetApp
	envResp, err := s.envVarService.ComputeAppEnvVars(ctx, db, &envvarservice.ComputeAppEnvVarsReq{
		App:  app,
		Sort: true,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	envVars := make([]string, 0, len(envResp))
	var errs []string
	for _, env := range envResp {
		envVars = append(envVars, env.ToString("="))
		errs = append(errs, env.Errors...)
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
