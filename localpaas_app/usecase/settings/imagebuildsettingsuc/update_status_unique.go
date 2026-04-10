package imagebuildsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) UpdateUniqueImageBuildSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.UpdateUniqueImageBuildSettingsStatusReq,
) (*imagebuildsettingsdto.UpdateUniqueImageBuildSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.UpdateUniqueImageBuildSettingsStatusResp{}, nil
}
