package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppContainerSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppContainerSettingsReq,
) (*appsettingsdto.GetAppContainerSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformContainerSettings(service)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppContainerSettingsResp{
		Data: resp,
	}, nil
}
