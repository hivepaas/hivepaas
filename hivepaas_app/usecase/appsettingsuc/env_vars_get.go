package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppEnvVarsReq,
) (*appsettingsdto.GetAppEnvVarsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ParentApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &appsettingsdto.EnvVarsTransformationInput{
		App:  app,
		Vars: settings,
	}
	input.SystemVars, err = uc.envVarService.ComputeAppSystemEnvVars(ctx, uc.db,
		&envvarservice.ComputeAppSystemEnvVarsReq{App: app})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ParentApp != nil {
		input.ParentSystemVars, err = uc.envVarService.ComputeAppSystemEnvVars(ctx, uc.db,
			&envvarservice.ComputeAppSystemEnvVarsReq{App: app.ParentApp})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	input.ProjectSystemVars, err = uc.envVarService.ComputeProjectSystemEnvVars(ctx, uc.db,
		&envvarservice.ComputeProjectSystemEnvVarsReq{Project: app.Project})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformEnvVars(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppEnvVarsResp{
		Data: resp,
	}, nil
}
