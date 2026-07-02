package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppReq,
) (*appdto.GetAppResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Tags",
			bunex.SelectOrder("display_order"),
		),
		bunex.SelectRelation("Settings",
			// NOTE: load http settings to extract active domain names of the app
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if app.ProjectID != req.ProjectID {
		return nil, apperrors.New(apperrors.ErrUnauthorized)
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats {
		serviceMap, err := uc.loadAppSwarmServices(ctx, app.Project.Key, []*entity.App{app})
		if err != nil {
			return nil, apperrors.New(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApp(app, transformationInput)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appdto.GetAppResp{
		Data: resp,
	}, nil
}
