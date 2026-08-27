package appplacementsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
)

func (uc *UC) DeleteAppPlacementSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appplacementsettingsdto.DeleteAppPlacementSettingsReq,
) (*appplacementsettingsdto.DeleteAppPlacementSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appplacementsettingsdto.DeleteAppPlacementSettingsResp{}, nil
}
