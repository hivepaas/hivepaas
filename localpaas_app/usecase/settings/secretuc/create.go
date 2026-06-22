package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
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
				_, err := uc.AppService.CreateSwarmSecret(ctx, db, data.ScopeApp, secret)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}

			err := pData.Setting.SetData(secret)
			if err != nil {
				return apperrors.Wrap(err)
			}
			pData.Setting.Size = secret.ValueSize()
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.CreateSecretResp{
		Data: resp.Data,
	}, nil
}
