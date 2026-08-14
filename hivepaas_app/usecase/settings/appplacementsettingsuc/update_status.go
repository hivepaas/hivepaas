package appplacementsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
)

func (uc *UC) UpdateAppPlacementSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *appplacementsettingsdto.UpdateAppPlacementSettingsStatusReq,
) (*appplacementsettingsdto.UpdateAppPlacementSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appplacementsettingsdto.UpdateAppPlacementSettingsStatusResp{}, nil
}
