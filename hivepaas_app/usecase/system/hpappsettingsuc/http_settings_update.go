package hpappsettingsuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) UpdateHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.UpdateHttpSettingsReq,
) (*hpappsettingsdto.UpdateHttpSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateHttpSettingsData{}
		err := uc.loadHttpSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingHttpSettings(ctx, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyHttpSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &hpappsettingsdto.UpdateHttpSettingsResp{}, nil
}

type updateHttpSettingsData struct {
	App             *entity.App
	HttpSetting     *entity.Setting
	NewHttpSettings *entity.AppHttpSettings
	RefObjects      *entity.RefObjects
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) loadHttpSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *hpappsettingsdto.UpdateHttpSettingsReq,
	data *updateHttpSettingsData,
) error {
	app, err := uc.appRepo.GetByGlobalKey(ctx, uc.db, "", base.HivepaasAppGlobalKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.HttpSetting = app.GetSettingByType(base.SettingTypeAppHttp)

	if data.HttpSetting != nil && data.HttpSetting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newHttpSettings := data.HttpSetting.MustAsAppHttpSettings()
	if err := req.ApplyTo(newHttpSettings); err != nil {
		return apperrors.Wrap(err)
	}
	data.NewHttpSettings = newHttpSettings

	// Make sure all reference settings used in these settings exist actively
	data.RefObjects, err = uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetObjectScope(),
		true, true, newHttpSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Active domains of the app need to validate
	activeDomains := newHttpSettings.GetActiveDomainNames()

	// Verify domains are allowed in project
	err = uc.domainService.VerifyProjectDomains(ctx, db, app.ProjectID, activeDomains)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Make sure all domains used by the app are not hold by any other app
	err = uc.domainService.VerifyDomainsAvailable(ctx, db, activeDomains, []string{app.ID})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) prepareUpdatingHttpSettings(
	_ context.Context,
	data *updateHttpSettingsData,
	persistingData *persistingAppData,
) {
	setting := data.HttpSetting
	timeNow := timeutil.NowUTC()

	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewHttpSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) applyHttpSettings(
	ctx context.Context,
	data *updateHttpSettingsData,
) error {
	appHttpSettings, err := data.HttpSetting.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	mapSslSettings := map[string]*entity.Setting{}
	for _, sslID := range appHttpSettings.GetSSLCertIDs() {
		if s := data.RefObjects.RefSettings[sslID]; s != nil {
			mapSslSettings[s.ID] = s
		}
	}
	err = uc.sslService.WriteCertFiles(false, gofn.MapValues(mapSslSettings)...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	inspect, err := uc.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service := &inspect.Service

	err = uc.traefikService.ApplyAppConfig(ctx, data.App, service, &traefikservice.AppConfigData{
		HttpSettings: appHttpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.networkService.UpdateAppGlobalRoutingNetwork(ctx, data.App, service, data.HttpSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
