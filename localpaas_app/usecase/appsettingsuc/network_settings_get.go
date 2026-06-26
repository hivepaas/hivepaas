package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppNetworkSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppNetworkSettingsReq,
) (*appsettingsdto.GetAppNetworkSettingsResp, error) {
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

	// TODO: query only networks used in the service
	listResp, err := uc.dockerManager.NetworkList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	networks := listResp.Items
	refObjects := &appsettingsdto.InfraRefObjects{
		Networks: make(map[string]*network.Summary, len(networks)),
	}
	for i := range networks { // faster than looping over structs?
		refObjects.Networks[networks[i].ID] = &networks[i]
	}

	resp, err := appsettingsdto.TransformNetworkSettings(service, refObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.GetAppNetworkSettingsResp{
		Data: resp,
	}, nil
}
