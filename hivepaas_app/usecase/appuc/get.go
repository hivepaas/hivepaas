package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppReq,
) (*appdto.GetAppResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("ParentApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Tags",
			bunex.SelectOrder("index"),
		),
		bunex.SelectRelation("Settings",
			// NOTE: load routing settings to extract active domain names of the app
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// NOTE: make sure we init the project env and project for the parent app
	if app.ParentApp != nil {
		app.ParentApp.Project = app.Project
		app.ParentApp.ProjectEnv = app.ProjectEnv
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats {
		serviceMap, err := uc.loadAppSwarmServices(ctx, app.Project.Key, []*entity.App{app})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApp(app, transformationInput)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdto.GetAppResp{
		Data: resp,
	}, nil
}
