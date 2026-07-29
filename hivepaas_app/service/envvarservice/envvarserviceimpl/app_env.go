package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

const (
	refSecretMaxSize = 10 * 1024 // 10 KB
)

func (s *service) ComputeEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeEnvVarsInAppReq,
) (*envvarservice.ComputeEnvVarsInAppResp, error) {
	envStore := make(map[string]*envvarservice.EnvVar, 20) //nolint:mnd
	secretStore := make(map[string]*entity.Setting, 10)    //nolint:mnd

	// Merge with inherited envs and secrets
	inheritedVars, inheritedSecrets, err := s.loadInheritedVarsAndSecretsInApp(ctx, db, req, envStore, secretStore)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Merge with envs of the current app
	currVars, currSecrets, err := s.loadVarsAndSecretsInApp(ctx, db, req.App, req.SkipLoadingVars,
		req.SkipLoadingSecrets, req.BuildPhaseOnly, req.OverridingVars)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	for _, aVar := range currVars {
		envStore[aVar.Key] = aVar
	}
	for _, aSec := range currSecrets {
		secretStore[aSec.Name] = aSec
	}

	refsData := &processRefsData{
		EnvStore:    envStore,
		SecretStore: secretStore,
		MaskSecrets: req.MaskSecrets,
		ExternalRefsLoadFunc: func(refName string) (map[string]*envvarservice.EnvVar, error) {
			resp, err := s.computeSharedEnvVarsInApp(ctx, db, req.App.ProjectID, req.App.ProjectEnvID,
				refName, req.BuildPhaseOnly, false, req.MaskSecrets)
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
		if req.BuildPhaseOnly && !env.IsBuild {
			continue
		}
		if req.SharedVarsOnly && !env.IsShared {
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

	if req.Sort {
		sort.Slice(resultVars, func(i, j int) bool {
			return resultVars[i].Key < resultVars[j].Key
		})
	}

	return &envvarservice.ComputeEnvVarsInAppResp{
		EnvVars:          resultVars,
		Secrets:          gofn.MapValues(secretStore),
		InheritedEnvVars: inheritedVars,
		InheritedSecrets: inheritedSecrets,
	}, nil
}

func (s *service) loadVarsAndSecretsInApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	skipLoadingVars bool,
	skipLoadingSecrets bool,
	buildPhase bool,
	overridingVars []*envvarservice.EnvVar,
) (envVars map[string]*envvarservice.EnvVar, secrets map[string]*entity.Setting, err error) {
	if skipLoadingVars && skipLoadingSecrets {
		return nil, nil, nil
	}

	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
			bunex.SelectWhereOrIf(!skipLoadingSecrets, "(setting.type = ? AND setting.size <= ?)",
				base.SettingTypeSecret, refSecretMaxSize),
		),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	envVars = make(map[string]*envvarservice.EnvVar, 20) //nolint:mnd
	secrets = make(map[string]*entity.Setting, 10)       //nolint:mnd
	for _, setting := range settings {
		if setting.Type == base.SettingTypeEnvVar {
			for _, env := range setting.MustAsEnvVars().Data {
				if env.IsBuild == buildPhase {
					envVars[env.Key] = &envvarservice.EnvVar{EnvVar: env}
				}
			}
		}
		if setting.Type == base.SettingTypeSecret {
			secrets[setting.Name] = setting
		}
	}

	// Inject overriding vars
	for _, aVar := range overridingVars {
		envVars[aVar.Key] = aVar
	}

	// Inject system env vars in app
	sysVars, err := s.ComputeSystemEnvVarsInApp(ctx, db, &envvarservice.ComputeSystemEnvVarsInAppReq{
		App: app,
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	for _, aVar := range sysVars {
		envVars[aVar.Key] = aVar
	}

	return envVars, secrets, nil
}

func (s *service) loadInheritedVarsAndSecretsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeEnvVarsInAppReq,
	envStore map[string]*envvarservice.EnvVar,
	secretStore map[string]*entity.Setting,
) (inheritedVars []*envvarservice.EnvVar, inheritedSecrets []*entity.Setting, err error) {
	// Merge with envs and secrets from parent project env
	inheritedVars, inheritedSecrets = req.InheritedVars, req.InheritedSecrets
	if inheritedVars == nil || inheritedSecrets == nil { //nolint:nestif
		if req.App.ParentApp != nil { // the app has a parent app, loads data from the parent
			req.App.ParentApp.ProjectEnv = req.App.ProjectEnv
			resp, err := s.ComputeEnvVarsInApp(ctx, db, &envvarservice.ComputeEnvVarsInAppReq{
				App:                req.App,
				SkipLoadingVars:    req.SkipLoadingVars,
				SkipLoadingSecrets: req.SkipLoadingSecrets,
				MaskSecrets:        req.MaskSecrets,
				BuildPhaseOnly:     req.BuildPhaseOnly,
				SharedVarsOnly:     req.SharedVarsOnly,
			})
			if err != nil {
				return nil, nil, apperrors.Wrap(err)
			}
			inheritedVars, inheritedSecrets = resp.EnvVars, resp.Secrets
		} else { // the app belongs to a project env directly, loads data from the env
			resp, err := s.ComputeEnvVarsInProjectEnv(ctx, db, &envvarservice.ComputeEnvVarsInProjectEnvReq{
				ProjectEnv:         req.App.ProjectEnv,
				SkipLoadingVars:    req.SkipLoadingVars,
				SkipLoadingSecrets: req.SkipLoadingSecrets,
				MaskSecrets:        req.MaskSecrets,
				BuildPhaseOnly:     req.BuildPhaseOnly,
				SharedVarsOnly:     req.SharedVarsOnly,
			})
			if err != nil {
				return nil, nil, apperrors.Wrap(err)
			}
			inheritedVars, inheritedSecrets = resp.EnvVars, resp.Secrets
		}
	}
	for _, aVar := range inheritedVars {
		envStore[aVar.Key] = aVar
	}
	for _, aSec := range inheritedSecrets {
		secretStore[aSec.Name] = aSec
	}

	if inheritedVars == nil {
		inheritedVars = []*envvarservice.EnvVar{}
	}
	if inheritedSecrets == nil {
		inheritedSecrets = []*entity.Setting{}
	}
	return inheritedVars, inheritedSecrets, nil
}
