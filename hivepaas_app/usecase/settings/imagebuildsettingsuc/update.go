package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	newSettings := req.ToEntity()
	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: string(currentSettingType),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			currSettings := pData.Setting.MustAsImageBuildSettings()
			if !newSettings.Sources.RepoCache && currSettings.Sources.RepoCache {
				_, _ = uc.forceClearRepoCache(ctx, db, req.Scope)
			}
			err := pData.Setting.SetData(newSettings)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// TODO (med): validate the IDs to be existing

	return &imagebuildsettingsdto.UpdateImageBuildSettingsResp{}, nil
}
