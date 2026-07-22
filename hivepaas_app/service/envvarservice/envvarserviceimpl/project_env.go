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
	// Trivial case
	hasRef := false
	for _, aVar := range req.TargetVars {
		if !aVar.IsLiteral && s.HasRef(aVar.Value) {
			hasRef = true
			break
		}
	}
	if !hasRef && len(req.TargetVars) > 0 {
		return gofn.MapSlice(req.TargetVars, func(v *entity.EnvVar) *envvarservice.EnvVar {
			return &envvarservice.EnvVar{EnvVar: v}
		}), nil
	}

	allVars, allSecrets, err := s.loadProjectVarsAndSecrets(ctx, db, req.Project, req.SkipLoadingVars,
		req.SkipLoadingSecrets, req.BuildPhaseOnly)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Merge target vars with the current
	for _, inVar := range req.TargetVars {
		allVars[inVar.Key] = &envvarservice.EnvVar{EnvVar: inVar}
	}

	refsData := &processRefsData{
		EnvStore:    allVars,
		SecretStore: allSecrets,
		MaskSecrets: req.MaskSecrets,
	}

	// Make a list of vars to compute
	resultVars := make([]*envvarservice.EnvVar, 0, len(allVars))
	if len(req.TargetVars) > 0 {
		for _, env := range req.TargetVars {
			if req.BuildPhaseOnly && !env.IsBuild {
				continue
			}
			resultVars = append(resultVars, &envvarservice.EnvVar{EnvVar: env})
		}
	} else {
		for _, env := range allVars {
			if req.BuildPhaseOnly && !env.IsBuild {
				continue
			}
			if req.SharedVarsOnly && !env.IsShared {
				continue
			}
			resultVars = append(resultVars, env)
		}
	}

	// Process all references within the ENV values
	for _, env := range resultVars {
		if !env.IsLiteral {
			err := s.processRefs(env, refsData)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
		}
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

	if len(settings) == 0 {
		return envVars, secrets, nil
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
