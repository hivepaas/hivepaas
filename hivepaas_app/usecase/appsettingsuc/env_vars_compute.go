package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ComputeAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.ComputeAppEnvVarsReq,
) (*appsettingsdto.ComputeAppEnvVarsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("ParentApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
		for _, env := range req.SharedEnvVars {
			envVars = append(envVars, &envvarservice.EnvVar{EnvVar: env.ToEntity(base.EnvVarKindShared)})
		}
	}

	computedVars, err := uc.envVarService.BuildEnvVarsInApp(ctx, uc.db, &envvarservice.BuildEnvVarsInAppReq{
		App:            app,
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
		return nil, hperrors.Wrap(err)
	}

	respEnvs := make([]*basedto.EnvVarResp, 0, len(computedVars.EnvVars))
	for _, env := range computedVars.EnvVars {
		respEnvs = append(respEnvs, basedto.TransformEnvVar(env.EnvVar))
	}

	return &appsettingsdto.ComputeAppEnvVarsResp{
		Data: respEnvs,
	}, nil
}
