package logginguc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc/loggingdto"
)

// GetSettings returns the stored configuration, or the disabled default when
// nothing has been saved, together with what is actually running.
func (uc *UC) GetSettings(
	ctx context.Context,
	_ *basedto.Auth,
	_ *loggingdto.GetSettingsReq,
) (*loggingdto.GetSettingsResp, error) {
	cfg, err := uc.loadCurrent(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	status, err := uc.loggingService.Status(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	data, masked := loggingdto.FromEntity(cfg)
	excluded := status.ExcludedApps
	if excluded == nil {
		excluded = []string{}
	}
	return &loggingdto.GetSettingsResp{
		Data:         data,
		SecretMasked: masked,
		Status:       &loggingdto.SettingsStatus{BackendReady: status.BackendReady, ExcludedApps: excluded},
	}, nil
}

// loadCurrent is the stored configuration, or nil when there is none yet.
func (uc *UC) loadCurrent(ctx context.Context) (*entity.Logging, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeLogging, true)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}
	cfg, err := setting.AsLogging()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return cfg, nil
}
