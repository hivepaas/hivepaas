package secretuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) DeleteSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.DeleteSecretReq,
) (*secretdto.DeleteSecretResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			if data.ScopeApp != nil {
				// Delete the related secret in docker swarm
				err := uc.ClusterService.DeleteSecretForApp(ctx, db, data.ScopeApp, data.Setting.MustAsSecret())
				if err != nil {
					return apperrors.Wrap(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.DeleteSecretResp{}, nil
}
