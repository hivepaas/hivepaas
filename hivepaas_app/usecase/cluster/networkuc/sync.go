package networkuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
)

func (uc *UC) SyncNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	_ *networkdto.SyncNetworkReq,
) (*networkdto.SyncNetworkResp, error) {
	_, err := uc.networkService.SyncNetworks(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &networkdto.SyncNetworkResp{}, nil
}
