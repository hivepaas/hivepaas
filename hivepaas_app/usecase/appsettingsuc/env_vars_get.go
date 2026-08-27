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

func (uc *UC) GetAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppEnvVarsReq,
) (*appsettingsdto.GetAppEnvVarsResp, error) {
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

	settings, _, err := uc.settingRepo.List(ctx, uc.db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appsettingsdto.EnvVarsTransformationInput{
		App:  app,
		Vars: settings,
	}
	input.SystemVars, err = uc.envVarService.BuildSystemEnvVarsInApp(ctx, uc.db,
		&envvarservice.BuildSystemEnvVarsInAppReq{App: app})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if app.ParentApp != nil {
		parentApp := app.ParentApp
		parentApp.Project = app.Project
		parentApp.ProjectEnv = app.ProjectEnv
		parentApp.ProjectEnv.Project = app.Project
		input.ParentSystemVars, err = uc.envVarService.BuildSystemEnvVarsInApp(ctx, uc.db,
			&envvarservice.BuildSystemEnvVarsInAppReq{App: parentApp})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	projectEnv := app.ProjectEnv
	projectEnv.Project = app.Project
	input.EnvSystemVars, err = uc.envVarService.BuildSystemEnvVarsInProjectEnv(ctx, uc.db,
		&envvarservice.BuildSystemEnvVarsInProjectEnvReq{ProjectEnv: projectEnv})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input.ProjectSystemVars, err = uc.envVarService.BuildSystemEnvVarsInProject(ctx, uc.db,
		&envvarservice.BuildSystemEnvVarsInProjectReq{Project: app.Project})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformEnvVars(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppEnvVarsResp{
		Data: resp,
	}, nil
}
