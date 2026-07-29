package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildEnvVarsInProject(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectReq,
) (*envvarservice.BuildEnvVarsInProjectResp, error) {
	envStore := make(map[string]*envvarservice.EnvVar, 20) //nolint:mnd
	secretStore := make(map[string]*entity.Setting, 10)    //nolint:mnd

	err := s.loadVarDataInProject(ctx, db, req, envStore, secretStore)
	if err != nil {
		return nil, apperrors.Wrap(err)
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

	return &envvarservice.BuildEnvVarsInProjectResp{
		EnvVars: resultVars,
		Secrets: gofn.MapValues(secretStore),
	}, nil
}

func (s *service) loadVarDataInProject(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildEnvVarsInProjectReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (err error) {
	project := req.Project
	dataLoadFunc := req.DataLoadFunc
	if dataLoadFunc == nil {
		dataLoadFunc = s.DefaultEnvLoad
	}

	loadedVars, loadedSecrets, err := dataLoadFunc(ctx, db, project.GetObjectScope(), req.LoadOptions)
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

	// Inject system env vars in project
	sysVars, err := s.BuildSystemEnvVarsInProject(ctx, db, &envvarservice.BuildSystemEnvVarsInProjectReq{
		Project: project,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, aVar := range sysVars {
		envStore[aVar.Key] = aVar
	}

	return nil
}
