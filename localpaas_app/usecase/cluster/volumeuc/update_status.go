package volumeuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
