package projectserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) ApplyEnvVarChangesToEnvs(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectReq,
) error {
	envVarData, err := s.envVarService.BuildEnvVarsInProject(ctx, db, req)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, projectEnv := range req.Project.ProjectEnvs {
		projectEnv.Project = req.Project
		envReq := &envvarservice.BuildEnvVarsInProjectEnvReq{
			ProjectEnv:            projectEnv,
			DataLoadFunc:          req.DataLoadFunc,
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(envVarData.EnvVars, envVarData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}

		// Run app update in a separate transaction to reduce lock time
		err = s.ExecuteEnvInTx(ctx, projectEnv, false, func(db database.Tx) error {
			if err := s.ApplyEnvVarChangesToApps(ctx, db, envReq); err != nil {
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

func (s *service) ApplyEnvVarChangesToApps(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectEnvReq,
) error {
	envVarData, err := s.envVarService.BuildEnvVarsInProjectEnv(ctx, db, req)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, app := range req.ProjectEnv.Apps {
		// If the app is a child of another app, ignore it, only process direct apps
		if app.IsChildApp() {
			continue
		}
		app.Project = req.ProjectEnv.Project
		app.ProjectEnv = req.ProjectEnv
		appReq := &envvarservice.BuildEnvVarsInAppReq{
			App:                   app,
			DataLoadFunc:          req.DataLoadFunc,
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(envVarData.EnvVars, envVarData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}

		// Run app update in a separate transaction to reduce lock time
		err = s.appService.ExecuteInTx(ctx, app, false, func(db database.Tx) error {
			if err := s.appService.ApplyEnvVarChanges(ctx, db, appReq); err != nil {
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
