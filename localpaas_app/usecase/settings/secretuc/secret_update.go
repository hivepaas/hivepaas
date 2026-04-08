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
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			oldSecret, err := pData.Setting.AsSecret()
			if err != nil {
				return apperrors.Wrap(err)
			}
			updatedSecret := req.ToEntity()
			if oldSecret != nil {
				updatedSecret.Key = oldSecret.Key // when update, keep the old KEY of the secret
			}
			if err = pData.Setting.SetData(updatedSecret); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.UpdateSecretResp{}, nil
}
