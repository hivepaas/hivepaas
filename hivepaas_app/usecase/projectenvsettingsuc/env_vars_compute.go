package projectenvsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

func (uc *UC) ComputeProjectEnvEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvsettingsdto.ComputeProjectEnvEnvVarsReq,
) (*projectenvsettingsdto.ComputeProjectEnvEnvVarsResp, error) {
	projectEnv, err := uc.projectEnvRepo.GetByID(ctx, uc.db, req.ProjectID, req.ProjectEnvID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	envVars := make([]*envvarservice.EnvVar, 0, 30) //nolint:mnd
	if len(req.BuildtimeEnvVars) > 0 {
		for _, env := range req.BuildtimeEnvVars {
			envVars = append(envVars, &envvarservice.EnvVar{EnvVar: env.ToEntity(base.EnvVarKindBuild)})
		}
	} else {
		for _, env := range req.RuntimeEnvVars {
			envVars = append(envVars, &envvarservice.EnvVar{EnvVar: env.ToEntity(base.EnvVarKindRuntime)})
		}
	}

	computedVars, err := uc.envVarService.BuildEnvVarsInProjectEnv(ctx, uc.db, &envvarservice.BuildEnvVarsInProjectEnvReq{
		ProjectEnv:     projectEnv,
		OverridingVars: envVars,
		LoadOptions: envvarservice.EnvLoadOptions{
			BuildPhase: len(req.BuildtimeEnvVars) > 0,
		},
		BuildOptions: envvarservice.EnvBuildOptions{
			BuildPhaseOnly: len(req.BuildtimeEnvVars) > 0,
			MaskSecrets:    true,
			Sort:           true,
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respEnvs := make([]*basedto.EnvVarResp, 0, len(computedVars.EnvVars))
	for _, env := range computedVars.EnvVars {
		respEnvs = append(respEnvs, basedto.TransformEnvVar(env.EnvVar))
	}

	return &projectenvsettingsdto.ComputeProjectEnvEnvVarsResp{
		Data: respEnvs,
	}, nil
}
