package logginguc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/logginguc/loggingdto"
)

const (
	currentSettingType = base.SettingTypeLogging
	loggingSettingName = "Logging settings"
)

// UpdateLoggingSettings stores the configuration and makes the cluster match it.
func (uc *UC) UpdateLoggingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *loggingdto.UpdateLoggingSettingsReq,
) (*loggingdto.UpdateLoggingSettingsResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	updateData := &updateSettingData{
		NewSettings: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: loggingSettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
		) error {
			updateData.UpdateUniqueSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			persistingData.PersistingSettingData = pData
			return uc.preparePersistingData(updateData, persistingData)
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			// Make the cluster match what was just saved. Apply is idempotent, so a
			// failure here leaves a stored configuration that the next save retries.
			if err := uc.loggingService.Apply(ctx, db); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &loggingdto.UpdateLoggingSettingsResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewSettings *entity.Logging
}

type persistingSettingData struct {
	*settings.PersistingSettingData
}

func (uc *UC) loadSettingData(
	ctx context.Context,
	db database.Tx,
	req *loggingdto.UpdateLoggingSettingsReq,
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
			Name:      loggingSettingName,
			Version:   entity.CurrentLoggingVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = setting

	currSettings, err := setting.AsLogging()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if currSettings == nil {
		currSettings = &entity.Logging{}
	}
	req.KeepMaskedSecrets(data.NewSettings, currSettings)

	if err := validateSettings(data.NewSettings); err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) preparePersistingData(
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	err := persistingData.Setting.SetData(updateData.NewSettings)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
