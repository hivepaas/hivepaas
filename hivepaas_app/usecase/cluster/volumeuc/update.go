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
	// NOTE: only allow updating `inheritable` and `default`
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.UpdateVolumeResp{}, nil
}
