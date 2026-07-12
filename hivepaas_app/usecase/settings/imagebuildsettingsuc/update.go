package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) UpdateImageBuildSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.UpdateImageBuildSettingsReq,
) (*imagebuildsettingsdto.UpdateImageBuildSettingsResp, error) {
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
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.UpdateImageBuildSettingsResp{}, nil
}
