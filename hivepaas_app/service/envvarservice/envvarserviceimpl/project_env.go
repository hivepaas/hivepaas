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

func (s *service) ComputeProjectEnvVars(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeProjectEnvVarsReq,
) ([]*envvarservice.EnvVar, error) {
	allVars, allSecrets, err := s.loadProjectVarsAndSecrets(ctx, db, req.Project, req.SkipLoadingVars,
		req.SkipLoadingSecrets, req.BuildPhaseOnly, req.OverridingVars)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	refsData := &processRefsData{
		EnvStore:    allVars,
		SecretStore: allSecrets,
		MaskSecrets: req.MaskSecrets,
	}

	resultVars := make([]*envvarservice.EnvVar, 0, len(allVars))
	var targetVarMap map[string]struct{}
	if len(req.TargetVars) > 0 {
		targetVarMap = gofn.MapSliceToMapKeys(req.TargetVars, struct{}{})
	}

	// Replace all references within the ENV values
	for _, env := range allVars {
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

func (s *service) loadProjectVarsAndSecrets(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
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
		bunex.SelectWhere("setting.object_id = ?", project.ID),
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

	// Inject project system env vars
	projectSysVars, err := s.ComputeProjectSystemEnvVars(ctx, &envvarservice.ComputeProjectSystemEnvVarsReq{
		Project: project,
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	for _, envVar := range projectSysVars {
		envVars[envVar.Key] = envVar
	}

	return envVars, secrets, nil
}
