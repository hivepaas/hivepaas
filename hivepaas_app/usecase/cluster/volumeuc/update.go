package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) UpdateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.UpdateVolumeReq,
) (*volumedto.UpdateVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth

	// Only `inheritable` and `default` are updatable. Everything that describes
	// the volume itself - where its data is, and how to mount it - is settled at
	// creation, because the data does not move when the description does.
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.UpdateVolumeResp{}, nil
}
