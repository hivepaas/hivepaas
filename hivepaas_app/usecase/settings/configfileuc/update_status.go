package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) UpdateConfigFileStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.UpdateConfigFileStatusReq,
) (*configfiledto.UpdateConfigFileStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		BeforePersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
			pData *settings.PersistingSettingStatusData,
		) (err error) {
			if req.Scope.App != nil {
				configFile := pData.Setting.MustAsConfigFile()
				if pData.Setting.IsActive() {
					// Create a config in docker swarm for the app
					_, err = uc.ClusterService.CreateConfigForApp(ctx, db, req.Scope.App, configFile)
				} else {
					// Delete the related config in docker swarm
					err = uc.ClusterService.DeleteConfigForApp(ctx, db, req.Scope.App, configFile)
				}
				if err != nil {
					return apperrors.Wrap(err)
				}
				pData.Setting.MustSetData(configFile)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &configfiledto.UpdateConfigFileStatusResp{}, nil
}
