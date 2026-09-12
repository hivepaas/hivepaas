package logginguc

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc/loggingdto"
)

// UpdateSettings stores the configuration and makes the cluster match it.
func (uc *UC) UpdateSettings(
	ctx context.Context,
	_ *basedto.Auth,
	req *loggingdto.UpdateSettingsReq,
) (*loggingdto.UpdateSettingsResp, error) {
	current, err := uc.loadCurrent(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := loggingdto.ToEntity(req.Data, current)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := validateSettings(cfg); err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	timeNow := timeutil.NowUTC()
	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.HivepaasScope,
			Type:      base.SettingTypeLogging,
			Status:    base.SettingStatusActive,
			Name:      "Logging",
			Version:   entity.CurrentLoggingVersion,
			CreatedAt: timeNow,
		}
	}
	setting.UpdatedAt = timeNow
	if err := setting.SetData(cfg); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := uc.settingRepo.Upsert(ctx, uc.db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Make the cluster match what was just saved. Apply is idempotent, so a
	// failure here leaves a stored configuration that the next save retries.
	if err := uc.loggingService.Apply(ctx, uc.db); err != nil {
		return nil, hperrors.Wrap(err)
	}

	data, masked := loggingdto.FromEntity(cfg)
	return &loggingdto.UpdateSettingsResp{Data: data, SecretMasked: masked}, nil
}
