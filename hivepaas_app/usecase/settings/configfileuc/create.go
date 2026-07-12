package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) CreateConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.CreateConfigFileReq,
) (*configfiledto.CreateConfigFileResp, error) {
	req.Type = currentSettingType
	configFile := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: configFile.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if data.ScopeApp != nil {
				// Create a config in docker swarm
				_, err := uc.ClusterService.CreateConfigForApp(ctx, db, data.ScopeApp, configFile)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}

			err := pData.Setting.SetData(configFile)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &configfiledto.CreateConfigFileResp{
		Data: resp.Data,
	}, nil
}
