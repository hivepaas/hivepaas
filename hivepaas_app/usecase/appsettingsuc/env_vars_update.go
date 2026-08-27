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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/envvarhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppEnvVarsReq,
) (*appsettingsdto.UpdateAppEnvVarsResp, error) {
	var buildData []*envvarservice.AppEnvVarData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppEnvVarsData{}
		err := uc.loadAppEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppEnvVars(req, data, persistingData)

		if !data.HasChanges {
			return nil
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Build and validate the env var changes
		buildData, err = uc.buildAppEnvVars(ctx, db, data, true)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &appsettingsdto.UpdateAppEnvVarsResp{
		Meta: &basedto.Meta{},
	}

	// Apply the changes to the app and child apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, uc.db, buildData, true, true)
	if len(errMap) == 0 {
		return resp, nil
	}
	// NOTE: just show user a message instead of failing the request?
	resp.Meta.Warning = "Configuration updated successfully, but failed to apply changes:"
	for i, e := range errMap {
		resp.Meta.Warning += fmt.Sprintf("\nApp '%v': ", buildData[i].App.Name) + e.Error()
	}

	return resp, nil
}

type updateAppEnvVarsData struct {
	App               *entity.App
	EnvVarsSetting    *entity.Setting
	CurrVars          []*entity.EnvVar
	HasChanges        bool
	RuntimeVarsChange bool
	SharedVarsChange  bool
	BuildVarsChange   bool
}

func (uc *UC) loadAppEnvVarsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppEnvVarsReq,
	data *updateAppEnvVarsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv.Apps"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app
	setting := app.GetSettingByType(base.SettingTypeEnvVar)
	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	data.EnvVarsSetting = setting
	if setting != nil {
		// Calculate current data to detect changes
		envVars, err := setting.AsEnvVars()
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.CurrVars = envVars.Data
	}

	return nil
}

func (uc *UC) prepareUpdatingAppEnvVars(
	req *appsettingsdto.UpdateAppEnvVarsReq,
	data *updateAppEnvVarsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.EnvVarsSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeApp,
			ObjectID:    app.ID,
			Type:        base.SettingTypeEnvVar,
			Inheritable: true,
			CreatedAt:   timeNow,
			Version:     entity.CurrentEnvVarsVersion,
		}
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive

	envVars := &entity.EnvVars{
		Data: make([]*entity.EnvVar, 0, len(req.RuntimeEnvVars)+len(req.SharedEnvVars)+len(req.BuildtimeEnvVars)),
	}
	for _, env := range req.SharedEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(base.EnvVarKindShared))
	}
	for _, env := range req.RuntimeEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(base.EnvVarKindRuntime))
	}
	for _, env := range req.BuildtimeEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(base.EnvVarKindBuild))
	}
	setting.MustSetData(envVars)

	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Detect changes in content
	data.RuntimeVarsChange, data.SharedVarsChange, data.BuildVarsChange =
		envvarhelper.CalcContentChanges(envVars.Data, data.CurrVars)
	// Detect any change even the order of the items
	data.HasChanges = data.RuntimeVarsChange || data.SharedVarsChange || data.BuildVarsChange
	if !data.HasChanges {
		data.HasChanges = envvarhelper.Equal(envVars.Data, data.CurrVars)
	}
}

func (uc *UC) buildAppEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateAppEnvVarsData,
	inTx bool,
) (appEnvVarData []*envvarservice.AppEnvVarData, err error) {
	scope := data.App.GetObjectScope()
	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	if data.SharedVarsChange {
		// When shared vars change, need to apply the changes to all apps in the project env
		// as some other apps may reference the vars of this app.
		projectEnv := data.App.ProjectEnv
		projectEnv.Project = data.App.Project
		appEnvVarData, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db,
			projectEnv.GetObjectScope(), false, nil, transaction, concurrency)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	} else if data.RuntimeVarsChange {
		appEnvVarData, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope,
			false, nil, transaction, concurrency)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if data.BuildVarsChange {
		// For build phase env vars, just validate them
		_, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope,
			true, nil, transaction, concurrency)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return appEnvVarData, nil
}
