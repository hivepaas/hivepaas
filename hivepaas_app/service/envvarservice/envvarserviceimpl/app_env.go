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

func (s *service) ComputeAppEnvVars(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeAppEnvVarsReq,
) ([]*envvarservice.EnvVar, error) {
	envStore := make(map[string]*envvarservice.EnvVar, 30) //nolint:mnd

	// Merge with inherited envs
	err := s.loadAppInheritedEnvVars(ctx, db, req, envStore)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Merge with envs of the current app
	appVars, appSecrets, err := s.loadAppVarsAndSecrets(ctx, db, req.App, req.SkipLoadingVars, req.SkipLoadingSecrets,
		req.BuildPhaseOnly, req.OverridingVars)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	for _, appVar := range appVars {
		envStore[appVar.Key] = appVar
	}

	refsData := &processRefsData{
		EnvStore:    envStore,
		SecretStore: appSecrets,
		MaskSecrets: req.MaskSecrets,
		ExternalRefsLoadFunc: func(refName string) (map[string]*envvarservice.EnvVar, error) {
			resp, err := s.computeAppSharedEnvVars(ctx, db, req.App.ProjectID, req.App.Env,
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

	return resultVars, nil
}

func (s *service) loadAppVarsAndSecrets(
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
	for _, env := range overridingVars {
		envVars[env.Key] = env
	}

	// Inject app system env vars
	sysVars, err := s.ComputeAppSystemEnvVars(ctx, db, &envvarservice.ComputeAppSystemEnvVarsReq{
		App: app,
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	for _, envVar := range sysVars {
		envVars[envVar.Key] = envVar
	}

	return envVars, secrets, nil
}

func (s *service) loadAppInheritedEnvVars(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeAppEnvVarsReq,
	envStore map[string]*envvarservice.EnvVar,
) (err error) {
	projectVars := req.InheritedProjectVars
	if projectVars == nil {
		projectVars, err = s.ComputeProjectEnvVars(ctx, db, &envvarservice.ComputeProjectEnvVarsReq{
			Project:            req.App.Project,
			SkipLoadingVars:    req.SkipLoadingVars,
			SkipLoadingSecrets: req.SkipLoadingSecrets,
			MaskSecrets:        req.MaskSecrets,
			BuildPhaseOnly:     req.BuildPhaseOnly,
			SharedVarsOnly:     req.SharedVarsOnly,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	for _, projectVar := range projectVars {
		envStore[projectVar.Key] = projectVar
	}

	// Merge with envs from parent app
	parentAppVars := req.InheritedParentAppVars
	if parentAppVars == nil && req.App.ParentID != "" {
		// Load parent app if not yet loaded
		if req.App.ParentApp == nil {
			req.App.ParentApp, err = s.appRepo.GetByID(ctx, db, req.App.ProjectID, req.App.ParentID,
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

		parentAppVars, err = s.ComputeAppEnvVars(ctx, db, &envvarservice.ComputeAppEnvVarsReq{
			App:                req.App.ParentApp,
			SkipLoadingVars:    req.SkipLoadingVars,
			SkipLoadingSecrets: req.SkipLoadingSecrets,
			MaskSecrets:        req.MaskSecrets,
			BuildPhaseOnly:     req.BuildPhaseOnly,
			SharedVarsOnly:     req.SharedVarsOnly,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	for _, parentVar := range parentAppVars {
		envStore[parentVar.Key] = parentVar
	}

	return nil
}
