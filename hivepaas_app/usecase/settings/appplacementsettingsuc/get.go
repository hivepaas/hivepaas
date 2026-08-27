package appplacementsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
)

func (uc *UC) GetAppPlacementSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appplacementsettingsdto.GetAppPlacementSettingsReq,
) (*appplacementsettingsdto.GetAppPlacementSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := appplacementsettingsdto.TransformImageBuild(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appplacementsettingsdto.GetAppPlacementSettingsResp{
		Data: respData,
	}, nil
}
