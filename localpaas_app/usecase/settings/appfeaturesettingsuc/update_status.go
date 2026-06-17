package appfeaturesettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
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
		return nil, apperrors.Wrap(err)
	}

	return &appfeaturesettingsdto.UpdateAppFeatureSettingsStatusResp{}, nil
}
