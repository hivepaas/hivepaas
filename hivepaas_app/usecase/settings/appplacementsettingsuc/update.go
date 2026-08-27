package appplacementsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
)

func (uc *UC) UpdateAppPlacementSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appplacementsettingsdto.UpdateAppPlacementSettingsReq,
) (*appplacementsettingsdto.UpdateAppPlacementSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: string(currentSettingType),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appplacementsettingsdto.UpdateAppPlacementSettingsResp{}, nil
}
