package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

const (
	refSecretMaxSize = 10 * 1024 // 10 KB
)

func (s *service) BuildEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInAppReq,
) (*envvarservice.BuildEnvVarsInAppResp, error) {
	envStore := make(map[string]*envvarservice.EnvVar, 30) //nolint:mnd
	secretStore := make(map[string]*entity.Setting, 20)    //nolint:mnd

	// Merge with inherited envs and secrets
	inheritedVars, inheritedSecrets, err := s.loadInheritedVarDataInApp(ctx, db, req, envStore, secretStore)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Merge with envs of the current app
	err = s.loadVarDataInApp(ctx, db, req, envStore, secretStore)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	refsData := &processRefsData{
		EnvStore:     envStore,
		SecretStore:  secretStore,
		BuildOptions: req.BuildOptions,
		ExternalRefsLoadFunc: func(refName string) (map[string]*envvarservice.EnvVar, error) {
			resp, err := s.buildSharedEnvVarsInApp(ctx, db, req.App.ProjectID, req.App.ProjectEnvID,
				refName, req.BuildOptions)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			respMap := make(map[string]*envvarservice.EnvVar, len(resp))
			for _, envVar := range resp {
				respMap[envVar.Key] = envVar
			}
			return respMap, nil
		},
	}

	resultVars := make([]*envvarservice.EnvVar, 0, len(envStore))
	var targetVarMap map[string]struct{}
	if len(req.TargetVars) > 0 {
		targetVarMap = gofn.MapSliceToMapKeys(req.TargetVars, struct{}{})
	}

	// Replace all references within the ENV values
	for _, env := range envStore {
		if req.BuildOptions.BuildPhaseOnly && !env.IsBuild {
			continue
		}
		if req.BuildOptions.SharedVarsOnly && !env.IsShared {
			continue
		}
		if targetVarMap != nil && !gofn.MapContainKeys(targetVarMap, env.Key) {
			continue
		}
		if !env.IsLiteral {
			if err = s.processRefs(env, refsData); err != nil {
				return nil, apperrors.Wrap(err)
			}
		}
		resultVars = append(resultVars, env)
	}

	if req.BuildOptions.Sort {
		sort.Slice(resultVars, func(i, j int) bool {
			return resultVars[i].Key < resultVars[j].Key
		})
	}

	return &envvarservice.BuildEnvVarsInAppResp{
		EnvVars:          resultVars,
		Secrets:          gofn.MapValues(secretStore),
		InheritedEnvVars: inheritedVars,
		InheritedSecrets: inheritedSecrets,
	}, nil
}

func (s *service) loadVarDataInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInAppReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (err error) {
	app := req.App
	dataLoadFunc := req.DataLoadFunc
	if dataLoadFunc == nil {
		dataLoadFunc = s.DefaultEnvLoad
	}

	loadedVars, loadedSecrets, err := dataLoadFunc(ctx, db, app.GetObjectScope(), envvarservice.EnvLoadOptions{
		BuildPhase: req.BuildOptions.BuildPhaseOnly,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, aVar := range loadedVars {
		envStore[aVar.Key] = aVar
	}
	for _, aSec := range loadedSecrets {
		secretStore[aSec.Name] = aSec
	}

	// Inject overriding vars
	for _, aVar := range req.OverridingVars {
		envStore[aVar.Key] = aVar
	}

	// Inject system env vars in app
	sysVars, err := s.BuildSystemEnvVarsInApp(ctx, db, &envvarservice.BuildSystemEnvVarsInAppReq{
		App: app,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, aVar := range sysVars {
		envStore[aVar.Key] = aVar
	}

	return nil
}

func (s *service) loadInheritedVarDataInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInAppReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (inheritedVars []*envvarservice.EnvVar, inheritedSecrets []*entity.Setting, err error) {
	app := req.App
	defaultLoadFunc := func(context.Context, database.IDB, *base.ObjectScope, envvarservice.EnvLoadOptions) (
		[]*envvarservice.EnvVar, []*entity.Setting, error) {
		app.ProjectEnv.Project = app.Project
		if app.ParentApp != nil { // the app has a parent app, loads data from the parent
			parentApp := app.ParentApp
			parentApp.ProjectEnv = app.ProjectEnv
			parentApp.Project = app.Project
			resp, err := s.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
				App:          parentApp,
				LoadOptions:  req.LoadOptions,
				BuildOptions: req.BuildOptions,
			})
			if err != nil {
				return nil, nil, apperrors.Wrap(err)
			}
			return resp.EnvVars, resp.Secrets, nil
		}

		// the app belongs to a project env directly, loads data from the env
		projectEnv := app.ProjectEnv
		projectEnv.Project = app.Project
		resp, err := s.BuildEnvVarsInProjectEnv(ctx, db, &envvarservice.BuildEnvVarsInProjectEnvReq{
			ProjectEnv:   projectEnv,
			LoadOptions:  req.LoadOptions,
			BuildOptions: req.BuildOptions,
		})
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		return resp.EnvVars, resp.Secrets, nil
	}

	loadFunc := req.InheritedDataLoadFunc
	if loadFunc == nil {
		loadFunc = defaultLoadFunc
	}

	inheritedVars, inheritedSecrets, err = loadFunc(ctx, db, app.GetObjectScope(), req.LoadOptions)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	for _, aVar := range inheritedVars {
		envStore[aVar.Key] = aVar
	}
	for _, aSec := range inheritedSecrets {
		secretStore[aSec.Name] = aSec
	}

	return inheritedVars, inheritedSecrets, nil
}
