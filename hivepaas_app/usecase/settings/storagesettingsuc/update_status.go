package storagesettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

func (uc *UC) UpdateStorageSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *storagesettingsdto.UpdateStorageSettingsStatusReq,
) (*storagesettingsdto.UpdateStorageSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &storagesettingsdto.UpdateStorageSettingsStatusResp{}, nil
}
