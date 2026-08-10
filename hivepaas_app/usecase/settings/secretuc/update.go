package secretuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) UpdateSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.UpdateSecretReq,
) (*secretdto.UpdateSecretResp, error) {
	req.Type = currentSettingType
	var oldSecret *entity.Secret
	updatedSecret := req.ToEntity()
	secretValueChanged := true
	var appEnvVarData []*envvarservice.AppEnvVarData
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingRefIDs: updatedSecret.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) (err error) {
			oldSecret, err = data.Setting.AsSecret()
			if err != nil {
				return apperrors.Wrap(err)
			}
			updatedSecret.Key = oldSecret.Key // when update, keep the old KEY of the secret
			if req.Value == "" {
				updatedSecret.Value = oldSecret.Value
				secretValueChanged = false
			}
			if err = pData.Setting.SetData(updatedSecret); err != nil {
				return apperrors.Wrap(err)
			}
			pData.Setting.Size, err = updatedSecret.ValueSize()
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) (err error) {
			// Rebuild affected env vars using the active transaction (inTx = true)
			if secretValueChanged {
				appEnvVarData, err = uc.buildAppEnvVarsForScope(ctx, db, req.Scope, true)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}
			if req.Scope.IsAppScope() {
				err = uc.ClusterSecretService.UpdateSecretForApp(ctx, db, req.Scope.App, oldSecret, updatedSecret)
				if err != nil {
					return apperrors.Wrap(err)
				}
				// Need to re-persist the setting as its content may change
				if updatedSecret.SwarmRef != nil && updatedSecret.SwarmRef.SecretID != "" {
					pData.Setting.MustSetData(updatedSecret)
					if err = uc.SettingRepo.Update(ctx, db, pData.Setting); err != nil {
						return apperrors.Wrap(err)
					}
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &secretdto.UpdateSecretResp{Meta: &basedto.Meta{}}

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

func (uc *UC) buildAppEnvVarsForScope(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	inTx bool,
) (appEnvVarData []*envvarservice.AppEnvVarData, err error) {
	switch scope.ScopeType {
	case base.ObjectScopeApp, base.ObjectScopeProject, base.ObjectScopeProjectEnv:
	case base.ObjectScopeUser, base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		fallthrough
	default:
		return nil, nil
	}

	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	// For build phase env vars, just validate them
	_, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope, true,
		nil, transaction, concurrency)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Runtime env vars
	if scope.IsProjectScope() || scope.IsProjectEnvScope() {
		appEnvVarData, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope, false,
			nil, transaction, concurrency)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return appEnvVarData, nil
	}

	// For app scope, changed secrets may cause shared vars to be changed, hence all other apps
	// reference this app may also need to be updated.
	// TODO: how to improve performance by calculating only the least change
	if scope.IsAppScope() {
		// Loads all apps in the env
		apps, _, err := uc.appRepo.List(ctx, db, scope.App.ProjectID, nil,
			bunex.SelectWhere("app.project_env_id = ?", scope.App.ProjectEnvID),
		)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		projectEnv := scope.App.ProjectEnv
		projectEnv.Project = scope.App.Project
		projectEnv.Apps = apps

		appEnvVarData, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db,
			projectEnv.GetObjectScope(), false, nil, transaction, concurrency)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		return appEnvVarData, nil
	}

	return nil, nil
}
