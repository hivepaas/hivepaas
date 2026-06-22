package configfileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc/configfiledto"
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
			if data.ScopeApp != nil {
				configFile := pData.Setting.MustAsConfigFile()
				if pData.Setting.IsActive() {
					// Create a config in docker swarm for the app
					_, err = uc.AppService.CreateSwarmConfig(ctx, db, data.ScopeApp, configFile)
				} else {
					// Delete the related config in docker swarm
					err = uc.AppService.DeleteSwarmConfig(ctx, db, data.ScopeApp, configFile)
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
