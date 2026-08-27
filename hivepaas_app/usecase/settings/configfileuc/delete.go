package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
			if req.Scope.App != nil {
				// Delete the related config in docker swarm
				err := uc.ClusterSecretService.RemoveConfigForApp(ctx, db, req.Scope.App, data.Setting.MustAsConfigFile())
				if err != nil {
					return hperrors.Wrap(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &configfiledto.DeleteConfigFileResp{}, nil
}
