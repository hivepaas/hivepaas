package envvarserviceimpl

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildEnvVarsForAllAppsInScope(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	buildPhase bool,
	transaction bool,
	concurrency bool,
) (appEnvVarData []*envvarservice.AppEnvVarData, err error) {
	switch scope.ScopeType {
	case base.ObjectScopeApp:
		appEnvVarData, err = s.BuildEnvVarsForAllAppsInApp(ctx, db,
			&envvarservice.BuildEnvVarsInAppReq{
				App: scope.App,
				LoadOptions: envvarservice.EnvLoadOptions{
					BuildPhase: buildPhase,
				},
				BuildOptions: envvarservice.EnvBuildOptions{
					BuildPhaseOnly: buildPhase,
				},
			}, transaction, concurrency)
	case base.ObjectScopeProjectEnv:
		appEnvVarData, err = s.BuildEnvVarsForAllAppsInProjectEnv(ctx, db,
			&envvarservice.BuildEnvVarsInProjectEnvReq{
				ProjectEnv: scope.ProjectEnv,
				LoadOptions: envvarservice.EnvLoadOptions{
					BuildPhase: buildPhase,
				},
				BuildOptions: envvarservice.EnvBuildOptions{
					BuildPhaseOnly: buildPhase,
				},
			}, transaction, concurrency)
	case base.ObjectScopeProject:
		appEnvVarData, err = s.BuildEnvVarsForAllAppsInProject(ctx, db,
			&envvarservice.BuildEnvVarsInProjectReq{
				Project: scope.Project,
				LoadOptions: envvarservice.EnvLoadOptions{
					BuildPhase: buildPhase,
				},
				BuildOptions: envvarservice.EnvBuildOptions{
					BuildPhaseOnly: buildPhase,
				},
			}, transaction, concurrency)
	case base.ObjectScopeUser, base.ObjectScopeGlobal:
		// Do nothing
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var errors []string
	for _, appData := range appEnvVarData {
		errors = append(errors, appData.Errors()...)
	}
	if len(errors) > 0 {
		return nil, apperrors.Wrap(apperrors.ErrValidation).WithDisplayLevelHigh().
			WithExtraDetail("%s", strings.Join(errors, "\n"))
	}

	return appEnvVarData, nil
}

func (s *service) BuildEnvVarsForAllAppsInProject(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectReq,
	transaction bool,
	concurrency bool,
) (result []*envvarservice.AppEnvVarData, err error) {
	projectData, err := s.BuildEnvVarsInProject(ctx, db, req)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, env := range req.Project.ProjectEnvs {
		env.Project = req.Project
		envReq := &envvarservice.BuildEnvVarsInProjectEnvReq{
			ProjectEnv:            env,
			DataLoadFunc:          req.DataLoadFunc,
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(projectData.EnvVars, projectData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}

		execFunc := func(ctx context.Context, db database.IDB) error {
			resp, err := s.BuildEnvVarsForAllAppsInProjectEnv(ctx, db, envReq, transaction, concurrency)
			if err != nil {
				return apperrors.Wrap(err)
			}
			result = append(result, resp...)
			return nil
		}

		if transaction {
			err = s.projectService.ExecuteEnvInTx(ctx, env, false, func(db database.Tx) error {
				return execFunc(ctx, db)
			})
		} else {
			err = execFunc(ctx, db)
		}
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return result, nil
}

func (s *service) BuildEnvVarsForAllAppsInProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectEnvReq,
	transaction bool,
	concurrency bool,
) (result []*envvarservice.AppEnvVarData, err error) {
	projectEnvData, err := s.BuildEnvVarsInProjectEnv(ctx, db, req)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(projectEnvData.EnvVars, projectEnvData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}

		execFunc := func(ctx context.Context, db database.IDB) error {
			resp, err := s.BuildEnvVarsForAllAppsInApp(ctx, db, appReq, transaction, concurrency)
			if err != nil {
				return apperrors.Wrap(err)
			}
			result = append(result, resp...)
			return nil
		}

		if transaction {
			err = s.appService.ExecuteInTx(ctx, app, false, func(db database.Tx) error {
				return execFunc(ctx, db)
			})
		} else {
			err = execFunc(ctx, db)
		}
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return result, nil
}

func (s *service) BuildEnvVarsForAllAppsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInAppReq,
	transaction bool,
	concurrency bool,
) (result []*envvarservice.AppEnvVarData, err error) {
	appData, err := s.BuildEnvVarsInApp(ctx, db, req)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	app := req.App
	result = append(result, &envvarservice.AppEnvVarData{
		App:     app,
		EnvVars: appData.EnvVars,
		Secrets: appData.Secrets,
	})

	if !app.IsChildApp() {
		return result, nil
	}

	for _, childApp := range app.ProjectEnv.GetChildAppsOfApp(app.ID) {
		childApp.ProjectEnv = app.ProjectEnv
		childApp.Project = app.Project
		childAppReq := &envvarservice.BuildEnvVarsInAppReq{
			App:                   childApp,
			DataLoadFunc:          req.DataLoadFunc,
			InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(appData.EnvVars, appData.Secrets),
			LoadOptions:           req.LoadOptions,
			BuildOptions:          req.BuildOptions,
		}

		execFunc := func(ctx context.Context, db database.IDB) error {
			resp, err := s.BuildEnvVarsInApp(ctx, db, childAppReq)
			if err != nil {
				return apperrors.Wrap(err)
			}
			result = append(result, &envvarservice.AppEnvVarData{
				App:     childApp,
				EnvVars: resp.EnvVars,
				Secrets: resp.Secrets,
			})
			return nil
		}

		if transaction {
			err = s.appService.ExecuteInTx(ctx, app, false, func(db database.Tx) error {
				return execFunc(ctx, db)
			})
		} else {
			err = execFunc(ctx, db)
		}
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return result, nil
}
