package traefiksettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

const (
	serviceUpdateMaxRetry = 2
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
			return hperrors.Wrap(err)
		}

		persistingData := &persistingSettingsData{}
		uc.prepareUpdatingServiceSettings(data, persistingData)

		err = uc.persistSettingsData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Before the service update, not after it. Changing the replica count
		// replaces traefik's tasks, and the connection carrying this request runs
		// through them - so a record that cannot be written has to be able to
		// stop the change rather than arrive after it.
		//
		// The replica counts go in with their values: they are numbers this
		// endpoint validates, nothing a user typed, and "who scaled traefik down
		// to one" is answered by the values or not at all.
		err = uc.recordTraefikSettingsUpdate(ctx, db, auth, auditSectionServiceSettings,
			auditdetail.New().Compare("replicas",
				data.CurrSettings.AppSettings.Replicas,
				data.NewSettings.AppSettings.Replicas))
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.traefikSvcChanges {
			err = uc.applyServiceSettingsToTraefikService(ctx, data)
			if err != nil {
				return hperrors.Wrap(err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &traefiksettingsdto.UpdateServiceSettingsResp{}, nil
}

type updateServiceSettingsData struct {
	Setting *entity.Setting
	// CurrSettings is the settings as they were before this request, kept so the
	// record can say what the values moved from.
	CurrSettings   *entity.TraefikService
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
		return hperrors.Wrap(err)
	}
	data.Setting = setting

	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	newSettings := req.ToEntity()
	data.NewSettings = newSettings

	currSettings, err := data.Setting.AsTraefikService()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.CurrSettings = currSettings

	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
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
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.TraefikService.ID, data.TraefikService,
		func(i int, svc *swarm.Service) (bool, error) {
			// Set service mode and replicas
			svc.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: new(uint64(data.NewSettings.AppSettings.Replicas)), //nolint:gosec
			}
			return true, nil
		}, serviceUpdateMaxRetry, 0)
	if err != nil {
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}
	return nil
}
