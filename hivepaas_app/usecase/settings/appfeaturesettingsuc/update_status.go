package appfeaturesettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

func (uc *UC) UpdateAppFeatureSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *appfeaturesettingsdto.UpdateAppFeatureSettingsStatusReq,
) (*appfeaturesettingsdto.UpdateAppFeatureSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appfeaturesettingsdto.UpdateAppFeatureSettingsStatusResp{}, nil
}
