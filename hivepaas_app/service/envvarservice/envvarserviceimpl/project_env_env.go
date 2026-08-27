package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildEnvVarsInProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectEnvReq,
) (*envvarservice.BuildEnvVarsInProjectEnvResp, error) {
	envStore := make(map[string]*envvarservice.EnvVar, 30) //nolint:mnd
	secretStore := make(map[string]*entity.Setting, 20)    //nolint:mnd

	// Merge with inherited envs and secrets
	inheritedVars, inheritedSecrets, err := s.loadInheritedVarDataInProjectEnv(ctx, db, req,
		envStore, secretStore)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Merge with envs and secrets of the current scope
	err = s.loadVarDataInProjectEnv(ctx, db, req, envStore, secretStore)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refsData := &processRefsData{
		EnvStore:     envStore,
		SecretStore:  secretStore,
		BuildOptions: req.BuildOptions,
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
				return nil, hperrors.Wrap(err)
			}
		}
		resultVars = append(resultVars, env)
	}

	if req.BuildOptions.Sort {
		sort.Slice(resultVars, func(i, j int) bool {
			return resultVars[i].Key < resultVars[j].Key
		})
	}

	return &envvarservice.BuildEnvVarsInProjectEnvResp{
		EnvVars:          resultVars,
		Secrets:          gofn.MapValues(secretStore),
		InheritedEnvVars: inheritedVars,
		InheritedSecrets: inheritedSecrets,
	}, nil
}

func (s *service) loadVarDataInProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectEnvReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (err error) {
	projectEnv := req.ProjectEnv
	dataLoadFunc := req.DataLoadFunc
	if dataLoadFunc == nil {
		dataLoadFunc = s.DefaultEnvLoad
	}

	loadedVars, loadedSecrets, err := dataLoadFunc(ctx, db, projectEnv.GetObjectScope(), req.LoadOptions)
	if err != nil {
		return hperrors.Wrap(err)
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

	// Inject system env vars in env
	sysVars, err := s.BuildSystemEnvVarsInProjectEnv(ctx, db, &envvarservice.BuildSystemEnvVarsInProjectEnvReq{
		ProjectEnv: projectEnv,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, aVar := range sysVars {
		envStore[aVar.Key] = aVar
	}

	return nil
}

func (s *service) loadInheritedVarDataInProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectEnvReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (inheritedVars []*envvarservice.EnvVar, inheritedSecrets []*entity.Setting, err error) {
	projectEnv := req.ProjectEnv
	defaultLoadFunc := func(context.Context, database.IDB, *entity.ObjectScope, envvarservice.EnvLoadOptions) (
		[]*envvarservice.EnvVar, []*entity.Setting, error) {
		resp, err := s.BuildEnvVarsInProject(ctx, db, &envvarservice.BuildEnvVarsInProjectReq{
			Project:      projectEnv.Project,
			LoadOptions:  req.LoadOptions,
			BuildOptions: req.BuildOptions,
		})
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		return resp.EnvVars, resp.Secrets, nil
	}

	loadFunc := req.InheritedDataLoadFunc
	if loadFunc == nil {
		loadFunc = defaultLoadFunc
	}

	inheritedVars, inheritedSecrets, err = loadFunc(ctx, db, projectEnv.GetObjectScope(), req.LoadOptions)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	for _, aVar := range inheritedVars {
		envStore[aVar.Key] = aVar
	}
	for _, aSec := range inheritedSecrets {
		secretStore[aSec.Name] = aSec
	}

	return inheritedVars, inheritedSecrets, nil
}
