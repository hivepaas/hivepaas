package registryuc

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryuc/registrydto"
)

const (
	currentSettingType  = base.SettingTypeRegistry
	registrySettingName = "Registry settings"
)

// UpdateRegistrySettings stores the configuration and makes the cluster match it.
func (uc *UC) UpdateRegistrySettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *registrydto.UpdateRegistrySettingsReq,
) (*registrydto.UpdateRegistrySettingsResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	updateData := &updateSettingData{NewSettings: req.ToEntity()}
	var applied *registryservice.SettingApplyResp

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: registrySettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
		) error {
			updateData.UpdateUniqueSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return hperrors.Wrap(pData.Setting.SetData(updateData.NewSettings))
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			// Make the cluster match what was just saved. Apply is idempotent, so
			// a failure here leaves a stored configuration the next save retries.
			applyResp, applyErr := uc.registryService.Apply(ctx, db, &registryservice.SettingApplyReq{
				Setting:       pData.Setting,
				TriggerUserID: auth.UserID(),
				Resources:     req.Resources(),
				RemoveApp:     req.RemoveApp,
				RemoveStorage: req.RemoveStorage,
			})
			applied = applyResp
			return hperrors.Wrap(applyErr)
		},
	})
	if err != nil {
		// The records went with the transaction; what provisioning made in docker
		// did not, and nothing else would take it down.
		if applied != nil && applied.Cleanup != nil {
			err = errors.Join(err, applied.Cleanup(context.WithoutCancel(ctx)))
		}
		return nil, hperrors.Wrap(err)
	}

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. The first deployment is what replaces the
	// placeholder image the app was created with; the certificate tasks are what
	// the domain needs.
	// Both are empty on every save but the one that provisioned the app.
	if applied != nil && applied.DeploymentTask != nil {
		if err = uc.taskQueue.ScheduleTask(ctx, applied.DeploymentTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	if applied != nil && len(applied.CertTasks) > 0 {
		if err = uc.taskQueue.ScheduleTask(ctx, applied.CertTasks...); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	result := &registrydto.UpdateRegistryRes{}
	if applied != nil && applied.RemovedApp {
		result.RemovedApp = true
		// After the commit, and through the settings usecase: it is what refuses
		// to delete a credential an app still names, and what records the
		// deletion. A refusal is the answer here rather than an error - the
		// registry is gone either way, and the account is still somebody's.
		kept, err := uc.removeCredential(ctx, auth, applied.RemovedCredentialID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		result.CredentialKept = kept
	}

	return &registrydto.UpdateRegistrySettingsResp{Data: result}, nil
}

// removeCredential deletes the account the registry was pushed with and says
// whether it had to be left alone.
func (uc *UC) removeCredential(ctx context.Context, auth *basedto.Auth, credentialID string) (bool, error) {
	if credentialID == "" {
		return false, nil
	}

	req := registryauthdto.NewDeleteRegistryAuthReq()
	req.ID = credentialID
	req.Scope = entity.NewObjectScopeGlobal()
	_, err := uc.registryAuthUC.DeleteRegistryAuth(ctx, auth, req)
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, hperrors.ErrSettingInUse):
		return true, nil
	case errors.Is(err, hperrors.ErrNotFound):
		// Already gone, which is the state this was asking for.
		return false, nil
	default:
		return false, hperrors.Wrap(err)
	}
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewSettings *entity.RegistrySettings
}

// loadSettingData reads what is stored, carries over what the client does not
// own, and refuses a configuration that could not work before anything is
// written.
func (uc *UC) loadSettingData(
	ctx context.Context,
	db database.Tx,
	req *registrydto.UpdateRegistrySettingsReq,
	data *updateSettingData,
) error {
	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, currentSettingType, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		timeNow := timeutil.NowUTC()
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType,
			Type:      currentSettingType,
			Status:    base.SettingStatusActive,
			Name:      registrySettingName,
			Version:   entity.CurrentRegistrySettingsVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = setting

	current, err := setting.AsRegistrySettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	// What provisioning wrote is the server's, not the client's: a request that
	// left them out would otherwise orphan the app and the credential.
	data.NewSettings.AppID = current.AppID
	data.NewSettings.RegistryAuthID = current.RegistryAuthID
	data.NewSettings.CredentialRotatedAt = current.CredentialRotatedAt
	data.NewSettings.CredentialGraceEnds = current.CredentialGraceEnds

	return hperrors.Wrap(uc.registryService.Validate(ctx, db, data.NewSettings, current))
}
