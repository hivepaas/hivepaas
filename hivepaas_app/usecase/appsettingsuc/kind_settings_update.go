package appsettingsuc

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppKindSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppKindSettingsReq,
) (*appsettingsdto.UpdateAppKindSettingsResp, error) {
	var data *updateAppKindSettingsData
	var persistingData *persistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateAppKindSettingsData{}
		err := uc.loadAppKindSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppKindSettings(ctx, req, data, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.postTxAppKindSettings(ctx, uc.db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppKindSettingsResp{}, nil
}

type updateAppKindSettingsData struct {
	App             *entity.App
	KindSetting     *entity.Setting
	NewKindSettings *entity.AppKindSettings
	RoutingSetting  *entity.Setting
	RefObjects      *entity.RefObjects
}

func (uc *UC) loadAppKindSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppKindSettingsReq,
	data *updateAppKindSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppKind, base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	kindSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind)
	routingSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting)

	if kindSetting != nil && kindSetting.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	var currKindSettings *entity.AppKindSettings
	if kindSetting != nil {
		currKindSettings = kindSetting.MustAsAppKindSettings()
	}
	newKindSettings := req.ToEntity()
	req.KeepMaskedSecrets(newKindSettings, currKindSettings)

	data.KindSetting = kindSetting
	data.NewKindSettings = newKindSettings
	data.RoutingSetting = routingSetting

	// Make sure all reference settings used in this settings exist actively
	refObjects := entity.NewRefObjects()
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, app.GetObjectScope(),
		true, newKindSettings.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.RefObjects = refObjects

	return nil
}

func (uc *UC) prepareUpdatingAppKindSettings(
	ctx context.Context,
	req *appsettingsdto.UpdateAppKindSettingsReq,
	data *updateAppKindSettingsData,
	persistingData *persistingAppData,
) error {
	app := data.App
	setting := data.KindSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppKind,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppKindSettingsVersion,
		}
		data.KindSetting = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive
	setting.MustSetData(data.NewKindSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Apply changes to routing settings
	err := uc.updateRoutingSettingsOnKindChange(ctx, req, data, persistingData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

//nolint:unparam
func (uc *UC) updateRoutingSettingsOnKindChange(
	_ context.Context,
	req *appsettingsdto.UpdateAppKindSettingsReq,
	data *updateAppKindSettingsData,
	persistingData *persistingAppData,
) error {
	app := data.App
	kindSettings := data.NewKindSettings
	routingSetting := data.RoutingSetting
	timeNow := timeutil.NowUTC()

	if routingSetting == nil {
		routingSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppRouting,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
			Version:   entity.CurrentAppRoutingSettingsVersion,
		}
		data.RoutingSetting = routingSetting
	}
	routingSettings := routingSetting.MustAsAppRoutingSettings()

	// Apply port change
	if int(req.Port) != routingSettings.Port {
		for _, domain := range routingSettings.Domains {
			if domain.ContainerPort == routingSettings.Port {
				domain.ContainerPort = int(req.Port)
			}
		}
		routingSettings.Port = int(req.Port)
	}

	hasPort := routingSettings.Port > 0
	var firstDomain *entity.AppDomain

	switch kindSettings.Category {
	case base.AppCategoryDatabase:
		if hasPort && req.Database.SSLCert.ID != "" {
			if len(routingSettings.Domains) == 0 {
				routingSettings.ExposePublicly = true
				firstDomain = &entity.AppDomain{
					Enabled:       true,
					ContainerPort: routingSettings.Port,
					Protocol:      base.NetworkProtocolTCP,
				}
				routingSettings.Domains = append(routingSettings.Domains, firstDomain)
			} else {
				firstDomain = routingSettings.Domains[0]
			}
			firstDomain.TLSPassthrough = req.Database.TLSPassthrough
			firstDomain.SSLCert = *req.Database.SSLCert.ToEntity()
		} else {
			routingSettings.ExposePublicly = false
			routingSettings.Domains = nil
		}
	case base.AppCategoryCache:
		if hasPort && req.Cache.SSLCert.ID != "" {
			if len(routingSettings.Domains) == 0 {
				routingSettings.ExposePublicly = true
				firstDomain = &entity.AppDomain{
					Enabled:       true,
					ContainerPort: routingSettings.Port,
					Protocol:      base.NetworkProtocolTCP,
				}
				routingSettings.Domains = append(routingSettings.Domains, firstDomain)
			} else {
				firstDomain = routingSettings.Domains[0]
			}
			firstDomain.SSLCert = *req.Cache.SSLCert.ToEntity()
		} else {
			routingSettings.ExposePublicly = false
			routingSettings.Domains = nil
		}
	case base.AppCategoryWebapp, base.AppCategoryStorage:
		// Their domains are the ordinary routing ones, changed on the routing
		// screen; nothing here decides them from the kind.
	}

	routingSetting.UpdateVer++
	routingSetting.UpdatedAt = timeNow
	routingSetting.ExpireAt = time.Time{}
	routingSetting.Status = base.SettingStatusActive
	routingSetting.MustSetData(routingSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, routingSetting)

	return nil
}

func (uc *UC) postTxAppKindSettings(
	ctx context.Context,
	db database.IDB,
	data *updateAppKindSettingsData,
) error {
	err := uc.applyAppRoutingSettingsOnKindChange(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.applyAppEnvVarsOnKindChange(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) applyAppRoutingSettingsOnKindChange(
	ctx context.Context,
	db database.IDB,
	data *updateAppKindSettingsData,
) error {
	routingSettings, err := data.RoutingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	_, err = uc.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             data.App,
		RoutingSettings: routingSettings,
		RefObjects:      data.RefObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) applyAppEnvVarsOnKindChange(
	ctx context.Context,
	db database.IDB,
	data *updateAppKindSettingsData,
) error {
	transaction := false
	concurrency := false

	// Loads all apps in the env
	apps, _, err := uc.appRepo.List(ctx, db, data.App.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", data.App.ProjectEnvID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	projectEnv := data.App.ProjectEnv
	projectEnv.Project = data.App.Project
	projectEnv.Apps = apps

	affectingAppEnvData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db,
		projectEnv.GetObjectScope(), false, nil, transaction, concurrency)
	if err != nil {
		return hperrors.Wrap(err)
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

	return hperrors.Wrap(hperrors.ErrActionFailed).WithExtraDetail("%s", warning)
}
