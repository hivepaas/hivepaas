package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppNetworkSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppNetworkSettingsReq,
) (*appsettingsdto.GetAppNetworkSettingsResp, error) {
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

	// List all networks in project env (app has no network scope)
	dbNets, dockerNets, err := uc.networkService.ListProjectEnvNetworks(ctx, uc.db, app.ProjectEnv)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appsettingsdto.NetworkTransformationInput{
		Service:        service,
		DBNetworks:     dbNets,
		DockerNetworks: dockerNets,
	}

	resp, err := appsettingsdto.TransformNetworkSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppNetworkSettingsResp{
		Data: resp,
	}, nil
}
