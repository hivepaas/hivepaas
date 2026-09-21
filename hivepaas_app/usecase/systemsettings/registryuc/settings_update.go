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
			_, applyErr := uc.registryService.Apply(ctx, db, &registryservice.SettingApplyReq{
				Setting:       pData.Setting,
				TriggerUserID: userIDOf(auth),
			})
			return hperrors.Wrap(applyErr)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registrydto.UpdateRegistrySettingsResp{}, nil
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

	return hperrors.Wrap(uc.registryService.Validate(data.NewSettings, current))
}

func userIDOf(auth *basedto.Auth) string {
	if auth == nil || auth.User == nil {
		return ""
	}
	return auth.User.ID
}
