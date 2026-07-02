package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) UpdateVolumeStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.UpdateVolumeStatusReq,
) (*volumedto.UpdateVolumeStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.UpdateVolumeStatusResp{}, nil
}
