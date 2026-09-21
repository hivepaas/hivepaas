package logginguc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/logginguc/loggingdto"
)

// GetLoggingSettings returns the stored configuration, or the disabled default when
// nothing has been saved, together with what is actually running.
func (uc *UC) GetLoggingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *loggingdto.GetLoggingSettingsReq,
) (*loggingdto.GetLoggingSettingsResp, error) {
	db := uc.DB
	req.Type = currentSettingType
	getResp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	setting := getResp.Data
	input := &loggingdto.LoggingSettingsTransformationInput{
		LoggingSetting: setting,
		RefObjects:     entity.NewRefObjects(),
		MaskSecrets:    !getResp.SecretsRevealed,
	}

	if setting != nil {
		status, err := uc.loggingService.Status(ctx, db, setting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		input.LoggingStatus = status

		err = uc.settingService.LoadRefObjectsSkipMissing(ctx, db, &input.RefObjects,
			entity.NewObjectScopeGlobal(), false, setting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	dataResp, err := loggingdto.TransformLoggingSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &loggingdto.GetLoggingSettingsResp{
		Data: dataResp,
	}, nil
}
