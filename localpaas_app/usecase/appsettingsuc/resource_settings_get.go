package appsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppResourceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppResourceSettingsReq,
) (*appsettingsdto.GetAppResourceSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := appsettingsdto.TransformResourceSettings(service)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.GetAppResourceSettingsResp{
		Data: resp,
	}, nil
}
