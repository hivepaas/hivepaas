package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/clusteraudit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
)

func (uc *UC) SyncVolume(
	ctx context.Context,
	auth *basedto.Auth,
	_ *volumedto.SyncVolumeReq,
) (*volumedto.SyncVolumeResp, error) {
	synced, err := uc.volumeService.SyncVolumes(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = clusteraudit.Record(ctx, uc.AuditService, uc.DB, auth,
		clusteraudit.Target{ResType: base.ResourceTypeClusterVolume}, "volume-sync",
		auditdetail.New().Set("syncedCount", len(synced)))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.SyncVolumeResp{}, nil
}
