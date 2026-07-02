package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) UpdateConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.UpdateConfigFileReq,
) (*configfiledto.UpdateConfigFileResp, error) {
	req.Type = currentSettingType
	updatedConfigFile := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingRefIDs: updatedConfigFile.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			oldConfigFile, err := pData.Setting.AsConfigFile()
			if err != nil {
				return apperrors.New(err)
			}
			if oldConfigFile != nil {
				updatedConfigFile.Name = oldConfigFile.Name // when update, keep the old NAME of the config
				if req.Content == "" {
					updatedConfigFile.Content = oldConfigFile.Content
				}
			}

			if data.ScopeApp != nil {
				// Update the related configs in docker swarm
				err := uc.ClusterService.UpdateConfigForApp(ctx, db, data.ScopeApp, oldConfigFile, updatedConfigFile)
				if err != nil {
					return apperrors.New(err)
				}
			}

			if err = pData.Setting.SetData(updatedConfigFile); err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &configfiledto.UpdateConfigFileResp{}, nil
}
