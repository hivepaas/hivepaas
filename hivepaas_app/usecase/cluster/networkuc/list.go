package networkuc

import (
	"context"

	"github.com/moby/moby/api/types/network"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) ListNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.ListNetworkReq,
) (_ *networkdto.ListNetworkResp, err error) {
	var currNets []network.Summary
	if req.Scope.IsGlobalScope() {
		currNets, err = uc.clusterService.SyncNetworks(ctx, uc.DB)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listNetworksInDocker(ctx, resp.Data, currNets, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := networkdto.TransformNetworks(resp.Data, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.ListNetworkResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}

func (uc *UC) listNetworksInDocker(
	ctx context.Context,
	settings []*entity.Setting,
	currNets []network.Summary,
	refClusterObjects *entity.RefClusterObjects,
) error {
	if currNets == nil {
		networks := make([]string, 0, len(settings))
		for _, setting := range settings {
			networks = append(networks, dockerhelper.ParseID(setting.ID))
		}
		if len(networks) == 0 {
			return nil
		}

		res, err := uc.dockerManager.NetworkListByIDs(ctx, networks)
		if err != nil {
			return apperrors.New(err)
		}
		currNets = res.Items
	}

	for i := range currNets {
		net := &currNets[i]
		refClusterObjects.RefNetworks[net.ID] = &net.Network
	}
	return nil
}
