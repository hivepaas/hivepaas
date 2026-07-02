package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) UpdateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.UpdateVolumeReq,
) (*volumedto.UpdateVolumeResp, error) {
	req.Type = currentSettingType
	// NOTE: only allow updating `availInProjects` and `default`
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.UpdateVolumeResp{}, nil
}
