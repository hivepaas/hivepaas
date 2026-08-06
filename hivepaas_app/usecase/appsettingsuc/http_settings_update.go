package appsettingsuc

import (
	"context"
	"fmt"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppHttpSettingsReq,
) (*appsettingsdto.UpdateAppHttpSettingsResp, error) {
	var data *updateAppHttpSettingsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateAppHttpSettingsData{}
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

	resp := &appsettingsdto.UpdateAppHttpSettingsResp{
		Meta: &basedto.Meta{},
	}

	if err = uc.applyAppEnvVars(ctx, uc.db, data, false); err != nil {
		// NOTE: just show user a message instead of failing the request?
		resp.Meta.Warning = "Configuration updated successfully, but failed to apply env var changes:\n" +
			err.Error()
	}

	return resp, nil
}

type updateAppHttpSettingsData struct {
	App             *entity.App
	HttpSetting     *entity.Setting
	NewHttpSettings *entity.AppHttpSettings
	RefObjects      *entity.RefObjects

	PortChanged   bool
	DomainChanged bool
}

func (uc *UC) loadAppHttpSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppHttpSettingsReq,
	data *updateAppHttpSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
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

	newHttpSettings := req.ToEntity()
	data.NewHttpSettings = newHttpSettings

	// Make sure all reference settings used in these settings exist actively
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &data.RefObjects, app.GetObjectScope(),
		true, newHttpSettings.GetRefObjectIDs())
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

	// Detect some changes
	var oldPort, newPort int
	var oldDomain, newDomain string
	if data.HttpSetting != nil {
		oldHttpSettings := data.HttpSetting.MustAsAppHttpSettings()
		oldPort = oldHttpSettings.Port
		if oldHttpSettings.ExposePublicly && len(oldHttpSettings.Domains) > 0 && oldHttpSettings.Domains[0].Enabled {
			oldDomain = oldHttpSettings.Domains[0].Domain
		}
	}
	newPort = newHttpSettings.Port
	if newHttpSettings.ExposePublicly && len(newHttpSettings.Domains) > 0 && newHttpSettings.Domains[0].Enabled {
		oldDomain = newHttpSettings.Domains[0].Domain
	}
	data.PortChanged = oldPort != newPort
	data.DomainChanged = oldDomain != newDomain

	return nil
}

func (uc *UC) prepareUpdatingAppHttpSettings(
	_ context.Context,
	data *updateAppHttpSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.HttpSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppHttp,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppHttpSettingsVersion,
		}
		data.HttpSetting = setting
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
	app := data.App
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

	inspect, err := uc.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service := &inspect.Service

	err = uc.traefikService.ApplyAppConfig(ctx, app, service, &traefikservice.AppConfigData{
		HttpSettings: appHttpSettings,
		RefObjects:   data.RefObjects,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.networkService.UpdateAppGlobalRoutingNetwork(ctx, app, service, data.HttpSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) applyAppEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateAppHttpSettingsData,
	inTx bool,
) error {
	if !data.PortChanged && !data.DomainChanged {
		return nil
	}

	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	// When port and domain change, we need to update env vars of the apps in the env
	// as port/domain are in shared env vars.

	// Loads all apps in the env
	apps, _, err := uc.appRepo.List(ctx, db, data.App.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", data.App.ProjectEnvID),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	projectEnv := data.App.ProjectEnv
	projectEnv.Project = data.App.Project
	projectEnv.Apps = apps

	affectingAppEnvData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db,
		projectEnv.GetObjectScope(), false, transaction, concurrency)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Apply the changes of env vars to the related apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, db, affectingAppEnvData, transaction, concurrency)
	if len(errMap) == 0 {
		return nil
	}

	var warning string
	for i, e := range errMap {
		warning += fmt.Sprintf("\nApp '%v': %v", affectingAppEnvData[i].App.Name, e.Error())
	}

	return apperrors.Wrap(apperrors.ErrActionFailed).WithExtraDetail("%s", warning)
}
