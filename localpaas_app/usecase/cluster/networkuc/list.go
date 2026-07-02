package networkuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

func (uc *UC) ListNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.ListNetworkReq,
) (*networkdto.ListNetworkResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listNetworksInDocker(ctx, resp.Data, refClusterObjects)
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
	refClusterObjects *entity.RefClusterObjects,
) error {
	networks := make([]string, 0, len(settings))
	for _, setting := range settings {
		net, err := setting.AsClusterNetwork()
		if err != nil {
			return apperrors.New(err)
		}
		networks = append(networks, net.NetworkID)
	}
	if len(networks) == 0 {
		return nil
	}

	if len(networks) == 1 {
		inspectResp, err := uc.dockerManager.NetworkInspect(ctx, networks[0])
		if err != nil {
			return apperrors.New(err)
		}
		refClusterObjects.RefNetworks[networks[0]] = &inspectResp.Network.Network
		return nil
	}

	res, err := uc.dockerManager.NetworkListByIDs(ctx, networks)
	if err != nil {
		return apperrors.New(err)
	}

	for i := range res.Items {
		net := &res.Items[i]
		refClusterObjects.RefNetworks[net.ID] = &net.Network
	}
	return nil
}
