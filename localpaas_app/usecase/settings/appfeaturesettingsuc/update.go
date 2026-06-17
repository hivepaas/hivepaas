package appfeaturesettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

func (uc *UC) UpdateAppFeatureSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appfeaturesettingsdto.UpdateAppFeatureSettingsReq,
) (*appfeaturesettingsdto.UpdateAppFeatureSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: "App feature settings",
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appfeaturesettingsdto.UpdateAppFeatureSettingsResp{}, nil
}
