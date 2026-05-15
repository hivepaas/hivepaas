package traefiksettingsuc

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

const (
	serviceUpdateMaxRetry      = 2
	serviceUpdateRetryInterval = time.Second * 3
)

func (uc *UC) UpdateServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.UpdateServiceSettingsReq,
) (*traefiksettingsdto.UpdateServiceSettingsResp, error) {
	var data *updateServiceSettingsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateServiceSettingsData{}
		err := uc.loadServiceSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingSettingsData{}
		uc.prepareUpdatingServiceSettings(data, persistingData)

		err = uc.persistSettingsData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.traefikSvcChanges {
			err = gofn.ExecRetry(func() error {
				return uc.applyServiceSettingsToTraefikService(ctx, data)
			}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &traefiksettingsdto.UpdateServiceSettingsResp{}, nil
}

type updateServiceSettingsData struct {
	Setting        *entity.Setting
	NewSettings    *entity.TraefikService
	TraefikService *swarm.Service

	traefikSvcChanges bool
}

func (uc *UC) loadServiceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *traefiksettingsdto.UpdateServiceSettingsReq,
	data *updateServiceSettingsData,
) error {
	setting, err := uc.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeTraefikService, true,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Setting = setting

	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newSettings := req.ToEntity()
	data.NewSettings = newSettings

	currSettings, err := data.Setting.AsTraefikService()
	if err != nil {
		return apperrors.Wrap(err)
	}

	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.TraefikService = traefikSvc

	if newSettings.AppSettings.Replicas != currSettings.AppSettings.Replicas {
		data.traefikSvcChanges = true
	}

	return nil
}

func (uc *UC) applyServiceSettingsToTraefikService(
	ctx context.Context,
	data *updateServiceSettingsData,
) error {
	traefikService := data.TraefikService

	// Set service mode and replicas
	traefikService.Spec.Mode.Replicated = &swarm.ReplicatedService{
		Replicas: new(uint64(data.NewSettings.AppSettings.Replicas)), //nolint:gosec
	}

	_, err := uc.dockerManager.ServiceUpdate(ctx, traefikService.ID, &traefikService.Version, &traefikService.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) prepareUpdatingServiceSettings(
	data *updateServiceSettingsData,
	persistingData *persistingSettingsData,
) {
	setting := data.Setting
	setting.MustSetData(data.NewSettings)
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	persistingData.Settings = append(persistingData.Settings, setting)
}

type persistingSettingsData struct {
	Settings []*entity.Setting
}

func (uc *UC) persistSettingsData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSettingsData,
) error {
	err := uc.settingRepo.UpsertMulti(ctx, db, persistingData.Settings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
