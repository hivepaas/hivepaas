package secretuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) UpdateSecretStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.UpdateSecretStatusReq,
) (*secretdto.UpdateSecretStatusResp, error) {
	req.Type = currentSettingType
	var appEnvVarData []*envvarservice.AppEnvVarData
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingStatusData,
			pData *settings.PersistingSettingStatusData,
		) (err error) {
			// Rebuild affected env vars using the active transaction (inTx = true)
			appEnvVarData, err = uc.buildAppEnvVarsForScope(ctx, db, req.Scope, true)
			if err != nil {
				return apperrors.Wrap(err)
			}

			if req.Scope.IsAppScope() {
				secret := pData.Setting.MustAsSecret()
				if pData.Setting.IsActive() {
					// Create a secret in the cluster for the app
					_, err = uc.ClusterService.CreateSecretForApp(ctx, db, req.Scope.App, secret)
				} else {
					// Delete the related secret in the cluster
					err = uc.ClusterService.DeleteSecretForApp(ctx, db, req.Scope.App, secret)
				}
				if err != nil {
					return apperrors.Wrap(err)
				}
				// Need to re-persist the setting as its content may change
				pData.Setting.MustSetData(secret)
				if err = uc.SettingRepo.Update(ctx, db, pData.Setting); err != nil {
					return apperrors.Wrap(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &secretdto.UpdateSecretStatusResp{Meta: &basedto.Meta{}}

	// Apply the changes to all affected apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, uc.DB, appEnvVarData, true, true)
	if len(errMap) == 0 {
		return resp, nil
	}
	// NOTE: just show user a message instead of failing the request?
	resp.Meta.Warning = "Secret updated successfully, but failed to apply changes to apps:"
	for i, e := range errMap {
		resp.Meta.Warning += fmt.Sprintf("\nApp '%v': ", appEnvVarData[i].App.Name) + e.Error()
	}

	return resp, nil
}
