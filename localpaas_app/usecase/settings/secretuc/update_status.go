package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) UpdateSecretStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.UpdateSecretStatusReq,
) (*secretdto.UpdateSecretStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		BeforePersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
			pData *settings.PersistingSettingStatusData,
		) (err error) {
			if data.ScopeApp != nil {
				secret := pData.Setting.MustAsSecret()
				if pData.Setting.IsActive() {
					// Create a secret in the cluster for the app
					_, err = uc.ClusterService.CreateSecretForApp(ctx, db, data.ScopeApp, secret)
				} else {
					// Delete the related secret in the cluster
					err = uc.ClusterService.DeleteSecretForApp(ctx, db, data.ScopeApp, secret)
				}
				if err != nil {
					return apperrors.New(err)
				}
				pData.Setting.MustSetData(secret)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &secretdto.UpdateSecretStatusResp{}, nil
}
