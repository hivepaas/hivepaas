package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) UpdateImageBuildSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.UpdateImageBuildSettingsStatusReq,
) (*imagebuildsettingsdto.UpdateImageBuildSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.UpdateImageBuildSettingsStatusResp{}, nil
}
