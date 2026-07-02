package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) DeleteConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.DeleteConfigFileReq,
) (*configfiledto.DeleteConfigFileResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			if data.ScopeApp != nil {
				// Delete the related config in docker swarm
				err := uc.ClusterService.DeleteConfigForApp(ctx, db, data.ScopeApp, data.Setting.MustAsConfigFile())
				if err != nil {
					return apperrors.New(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &configfiledto.DeleteConfigFileResp{}, nil
}
