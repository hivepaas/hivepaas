package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) DeleteImageBuildSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.DeleteImageBuildSettingsReq,
) (*imagebuildsettingsdto.DeleteImageBuildSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.DeleteImageBuildSettingsResp{}, nil
}
