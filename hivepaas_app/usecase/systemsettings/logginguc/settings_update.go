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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
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
	var applied *loggingservice.SettingApplyResp

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
			applyResp, applyErr := uc.loggingService.Apply(ctx, db, &loggingservice.SettingApplyReq{
				Setting:          pData.Setting,
				BackendResources: req.BackendResources(),
				TriggerUserID:    auth.UserID(),
				RemoveApp:        req.RemoveApp,
				RemoveStorage:    req.RemoveStorage,
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
	// transaction has committed. The deployments are what replace a new app's
	// placeholder image, and what apply a changed configuration to one running.
	if applied != nil && len(applied.Tasks) > 0 {
		if err = uc.taskQueue.ScheduleTask(ctx, applied.Tasks...); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &loggingdto.UpdateLoggingSettingsResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewSettings *entity.LoggingSettings
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
			Version:   entity.CurrentLoggingSettingsVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = setting

	currSettings, err := setting.AsLoggingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	req.KeepMaskedSecrets(data.NewSettings, currSettings)
	// What provisioning wrote is the server's, not the client's: a request that
	// left them out would otherwise drop the links to the apps.
	data.NewSettings.BackendAppID = currSettings.BackendAppID
	data.NewSettings.CollectorAppID = currSettings.CollectorAppID

	if err := validateSettings(data.NewSettings); err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.loggingService.Validate(ctx, db, data.NewSettings, currSettings))
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
