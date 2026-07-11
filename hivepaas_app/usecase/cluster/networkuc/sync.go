package networkuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
)

func (uc *UC) SyncNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	_ *networkdto.SyncNetworkReq,
) (*networkdto.SyncNetworkResp, error) {
	_, err := uc.clusterService.SyncNetworks(ctx, uc.DB)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return &networkdto.SyncNetworkResp{}, nil
}
