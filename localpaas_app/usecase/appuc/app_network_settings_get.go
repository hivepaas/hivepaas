package appuc

import (
	"context"

	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetAppNetworkSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppNetworkSettingsReq,
) (*appdto.GetAppNetworkSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// TODO: query only networks used in the service
	networks, err := uc.dockerManager.NetworkList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	refObjects := &appdto.InfraRefObjects{
		Networks: make(map[string]*network.Summary, len(networks)),
	}
	for i := range networks { // faster than looping over structs?
		refObjects.Networks[networks[i].ID] = &networks[i]
	}

	resp, err := appdto.TransformNetworkSettings(service, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.GetAppNetworkSettingsResp{
		Data: resp,
	}, nil
}
