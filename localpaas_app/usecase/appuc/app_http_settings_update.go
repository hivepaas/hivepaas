package appuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppHttpSettingsReq,
) (*appdto.UpdateAppHttpSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppHttpSettingsData{}
		err := uc.loadAppHttpSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppHttpSettings(ctx, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppHttpSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppHttpSettingsResp{}, nil
}

type updateAppHttpSettingsData struct {
	App             *entity.App
	HttpSettings    *entity.Setting
	NewHttpSettings *entity.AppHttpSettings
	RefObjects      *entity.RefObjects
}

func (uc *UC) loadAppHttpSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppHttpSettingsReq,
	data *updateAppHttpSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.HttpSettings, _ = gofn.First(app.Settings)

	if data.HttpSettings != nil && data.HttpSettings.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newHttpSettings := req.ToEntity()
	data.NewHttpSettings = newHttpSettings

	// Make sure all reference settings used in this settings exist actively
	data.RefObjects, err = uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetSettingScope(),
		true, true, newHttpSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppHttpSettings(
	_ context.Context,
	data *updateAppHttpSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.HttpSettings
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.SettingScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppHttp,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppHttpSettingsVersion,
		}
		data.HttpSettings = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewHttpSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) applyAppHttpSettings(
	ctx context.Context,
	data *updateAppHttpSettingsData,
) error {
	appHttpSettings, err := data.HttpSettings.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	sslSettings := map[string]*entity.Setting{}
	for _, sslID := range appHttpSettings.GetSSLCertIDs() {
		if s := data.RefObjects.RefSettings[sslID]; s != nil {
			sslSettings[s.ID] = s
		}
	}
	err = uc.sslService.WriteCertFiles(false, gofn.MapValues(sslSettings)...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	service, err := uc.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.traefikService.ApplyAppConfig(ctx, data.App, service, &traefikservice.AppConfigData{
		HttpSettings: appHttpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.networkService.UpdateAppGlobalRoutingNetwork(ctx, data.App, service, data.HttpSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
