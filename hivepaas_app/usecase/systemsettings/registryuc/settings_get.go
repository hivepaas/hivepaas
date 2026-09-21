package registryuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryuc/registrydto"
)

// GetRegistrySettings returns the stored configuration, or the disabled default
// when nothing has been saved, together with what is actually running.
func (uc *UC) GetRegistrySettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *registrydto.GetRegistrySettingsReq,
) (*registrydto.GetRegistrySettingsResp, error) {
	db := uc.DB
	req.Type = currentSettingType
	getResp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq,
		&settings.GetUniqueSettingData{})
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	setting := getResp.Data
	input := &registrydto.RegistrySettingsTransformationInput{RegistrySetting: setting}
	if setting != nil {
		status, statusErr := uc.registryService.Status(ctx, db, setting)
		if statusErr != nil {
			return nil, hperrors.Wrap(statusErr)
		}
		input.RegistryStatus = status
	}

	dataResp, err := registrydto.TransformRegistrySettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registrydto.GetRegistrySettingsResp{Data: dataResp}, nil
}
