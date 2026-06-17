package appfeaturesettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

func (uc *UC) DeleteAppFeatureSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appfeaturesettingsdto.DeleteAppFeatureSettingsReq,
) (*appfeaturesettingsdto.DeleteAppFeatureSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appfeaturesettingsdto.DeleteAppFeatureSettingsResp{}, nil
}
