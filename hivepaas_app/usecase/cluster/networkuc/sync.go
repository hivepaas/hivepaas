package networkuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/clusteraudit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
)

func (uc *UC) SyncNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	_ *networkdto.SyncNetworkReq,
) (*networkdto.SyncNetworkResp, error) {
	synced, err := uc.networkService.SyncNetworks(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = clusteraudit.Record(ctx, uc.AuditService, uc.DB, auth,
		clusteraudit.Target{ResType: base.ResourceTypeClusterNetwork}, "network-sync",
		auditdetail.New().Set("syncedCount", len(synced)))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &networkdto.SyncNetworkResp{}, nil
}
