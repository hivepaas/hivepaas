package networkuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

func (uc *UC) GetNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.GetNetworkReq,
) (*networkdto.GetNetworkResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listNetworksInDocker(ctx, []*entity.Setting{resp.Data}, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := networkdto.TransformNetwork(resp.Data, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.GetNetworkResp{
		Data: respData,
	}, nil
}
