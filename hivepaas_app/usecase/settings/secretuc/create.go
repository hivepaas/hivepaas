package secretuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) CreateSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.CreateSecretReq,
) (*secretdto.CreateSecretResp, error) {
	req.Type = currentSettingType
	secret := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Key,
		VerifyingRefIDs: secret.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if data.ScopeApp != nil {
				// Create a secret in docker swarm
				_, err := uc.ClusterService.CreateSecretForApp(ctx, db, data.ScopeApp, secret)
				if err != nil {
					return apperrors.New(err)
				}
			}

			err := pData.Setting.SetData(secret)
			if err != nil {
				return apperrors.New(err)
			}
			pData.Setting.Size, err = secret.ValueSize()
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &secretdto.CreateSecretResp{
		Data: resp.Data,
	}, nil
}
