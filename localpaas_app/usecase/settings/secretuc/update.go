package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) UpdateSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.UpdateSecretReq,
) (*secretdto.UpdateSecretResp, error) {
	req.Type = currentSettingType
	updatedSecret := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingRefIDs: updatedSecret.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			oldSecret, err := pData.Setting.AsSecret()
			if err != nil {
				return apperrors.New(err)
			}
			if oldSecret != nil {
				updatedSecret.Key = oldSecret.Key // when update, keep the old KEY of the secret
				if req.Value == "" {
					updatedSecret.Value = oldSecret.Value
				}
			}

			if data.ScopeApp != nil {
				// Update the related secrets in docker swarm
				err := uc.ClusterService.UpdateSecretForApp(ctx, db, data.ScopeApp, oldSecret, updatedSecret)
				if err != nil {
					return apperrors.New(err)
				}
			}

			if err = pData.Setting.SetData(updatedSecret); err != nil {
				return apperrors.New(err)
			}
			pData.Setting.Size, err = updatedSecret.ValueSize()
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &secretdto.UpdateSecretResp{}, nil
}
