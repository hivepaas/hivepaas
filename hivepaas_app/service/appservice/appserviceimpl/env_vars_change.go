package appserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) ApplyEnvVarChanges(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInAppReq,
) error {
	req.BuildOptions.Sort = true
	envVarData, err := s.envVarService.BuildEnvVarsInApp(ctx, db, req)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if !req.BuildOptions.BuildPhaseOnly {
		if err = s.applyEnvVarChangesToService(ctx, req, envVarData); err != nil {
			return apperrors.Wrap(err)
		}
	}

	// Apply changes of env vars in child apps (preview apps)
	if req.App.IsChildApp() {
		return nil
	}

	app := req.App
	var childApps []*entity.App
	if app.ProjectEnv != nil && len(app.ProjectEnv.Apps) > 0 {
		childApps = app.ProjectEnv.GetChildAppsOfApp(app.ID)
	} else {
		childApps, _, err = s.appRepo.List(ctx, db, "", nil,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	for _, childApp := range childApps {
		childApp.ProjectEnv = app.ProjectEnv
		childApp.Project = app.Project
		childAppReq := &envvarservice.BuildEnvVarsInAppReq{
			App:                   childApp,
			DataLoadFunc:          req.DataLoadFunc,
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(envVarData.EnvVars, envVarData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}
		err = s.ExecuteInTx(ctx, childApp, false, func(db database.Tx) error {
			if err := s.ApplyEnvVarChanges(ctx, db, childAppReq); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) applyEnvVarChangesToService(
	ctx context.Context,
	req *envvarservice.BuildEnvVarsInAppReq,
	envVarData *envvarservice.BuildEnvVarsInAppResp,
) error {
	if req.BuildOptions.BuildPhaseOnly {
		return nil
	}

	app := req.App
	envVars := make([]string, 0, len(envVarData.EnvVars))
	var errors []string
	for _, env := range envVarData.EnvVars {
		envVars = append(envVars, env.ToString("="))
		for _, e := range env.Errors {
			errors = append(errors, e.ErrorWithApp(app.Name))
		}
	}
	if len(errors) > 0 {
		return apperrors.Wrap(apperrors.ErrEnvVarContainInvalidReference).WithDisplayLevelHigh().
			WithExtraDetail("%s", strings.Join(errors, "\n"))
	}

	service, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if service.Spec.TaskTemplate.ContainerSpec == nil {
		service.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}

	currEnvVars := service.Spec.TaskTemplate.ContainerSpec.Env
	if gofn.ContentEqual(currEnvVars, envVars) { // No change, just return
		return nil
	}

	service.Spec.TaskTemplate.ContainerSpec.Env = envVars
	_, err = s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
