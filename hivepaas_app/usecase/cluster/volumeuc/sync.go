package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
)

func (uc *UC) SyncVolume(
	ctx context.Context,
	auth *basedto.Auth,
	_ *volumedto.SyncVolumeReq,
) (*volumedto.SyncVolumeResp, error) {
	_, err := uc.clusterService.SyncNetworks(ctx, uc.DB)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.SyncVolumeResp{}, nil
}
