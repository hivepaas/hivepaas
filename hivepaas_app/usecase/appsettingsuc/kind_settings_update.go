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
		uc.prepareUpdatingAppKindSettings(req, data, persistingData)

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
	RoutingChanged  bool
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

	// Make sure all reference settings used in this settings exist actively. A
	// port change applies the routing settings again, which needs what they
	// refer to as well: their certificates, their basic auth.
	refIDs := newKindSettings.GetRefObjectIDs()
	if routingSetting != nil {
		routingSettings := routingSetting.MustAsAppRoutingSettings()
		if int(req.Port) != routingSettings.Port {
			refIDs.AddRefIDs(routingSettings.GetRefObjectIDs())
		}
	}
	refObjects := entity.NewRefObjects()
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, app.GetObjectScope(),
		true, refIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.RefObjects = refObjects

	return nil
}

func (uc *UC) prepareUpdatingAppKindSettings(
	req *appsettingsdto.UpdateAppKindSettingsReq,
	data *updateAppKindSettingsData,
	persistingData *persistingAppData,
) {
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

	// The app's port is part of what it is, and lives in its routing settings.
	uc.updateRoutingPortOnKindChange(req, data, persistingData)
}

// updateRoutingPortOnKindChange moves the app's port, and its domains' that
// pointed at it, when the kind settings change it. It is all the kind settings
// do to routing: domains, their certificates and TLS passthrough belong to the
// routing settings alone - the kind settings used to write them too, and
// wiped a database's domains whenever it was saved without a certificate
// picked, as with one still being obtained.
func (uc *UC) updateRoutingPortOnKindChange(
	req *appsettingsdto.UpdateAppKindSettingsReq,
	data *updateAppKindSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	routingSetting := data.RoutingSetting
	timeNow := timeutil.NowUTC()

	var routingSettings *entity.AppRoutingSettings
	if routingSetting != nil {
		routingSettings = routingSetting.MustAsAppRoutingSettings()
	} else {
		routingSettings = &entity.AppRoutingSettings{}
	}
	if int(req.Port) == routingSettings.Port {
		return
	}

	for _, domain := range routingSettings.Domains {
		if domain.ContainerPort == routingSettings.Port {
			domain.ContainerPort = int(req.Port)
		}
	}
	routingSettings.Port = int(req.Port)

	if routingSetting == nil {
		routingSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppRouting,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppRoutingSettingsVersion,
		}
		data.RoutingSetting = routingSetting
	}
	routingSetting.UpdateVer++
	routingSetting.UpdatedAt = timeNow
	routingSetting.ExpireAt = time.Time{}
	routingSetting.Status = base.SettingStatusActive
	routingSetting.MustSetData(routingSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, routingSetting)
	data.RoutingChanged = true
}

func (uc *UC) postTxAppKindSettings(
	ctx context.Context,
	db database.IDB,
	data *updateAppKindSettingsData,
) error {
	if data.RoutingChanged {
		err := uc.applyAppRoutingSettingsOnKindChange(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	err := uc.applyAppEnvVarsOnKindChange(ctx, db, data)
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
