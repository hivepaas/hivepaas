package projectsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) ComputeProjectEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.ComputeProjectEnvVarsReq,
) (*projectsettingsdto.ComputeProjectEnvVarsResp, error) {
	project, err := uc.projectRepo.GetByID(ctx, uc.db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
	)
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

	computedVars, err := uc.envVarService.ComputeProjectEnvVars(ctx, uc.db, &envvarservice.ComputeProjectEnvVarsReq{
		Project:        project,
		OverridingVars: envVars,
		BuildPhaseOnly: len(req.BuildtimeEnvVars) > 0,
		MaskSecrets:    true,
		Sort:           true,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respEnvs := make([]*basedto.EnvVarResp, 0, len(computedVars))
	for _, env := range computedVars {
		respEnvs = append(respEnvs, basedto.TransformEnvVar(env.EnvVar))
	}

	return &projectsettingsdto.ComputeProjectEnvVarsResp{
		Data: respEnvs,
	}, nil
}
