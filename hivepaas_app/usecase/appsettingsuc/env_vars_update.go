package appsettingsuc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppEnvVars(req, data, persistingData)

		if !data.BuildVarsChange && !data.RuntimeVarsChange {
			return nil
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Build and validate the env var changes
		buildData, err = uc.buildAppEnvVars(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
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
	App            *entity.App
	EnvVarsSetting *entity.Setting

	CurrRuntimeVars []*entity.EnvVar
	CurrBuildVars   []*entity.EnvVar

	RuntimeVarsChange bool
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
		return apperrors.Wrap(err)
	}
	data.App = app
	setting := app.GetSettingByType(base.SettingTypeEnvVar)
	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	data.RuntimeVarsChange = true
	data.BuildVarsChange = true
	data.EnvVarsSetting = setting

	if setting == nil {
		return nil
	}

	// Calculate current data to detect changes
	envVars, err := setting.AsEnvVars()
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, env := range envVars.Data {
		if env.IsBuild {
			data.CurrBuildVars = append(data.CurrBuildVars, env)
		} else {
			data.CurrRuntimeVars = append(data.CurrRuntimeVars, env)
		}
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
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeEnvVar,
			CreatedAt: timeNow,
			Version:   entity.CurrentEnvVarsVersion,
		}
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive

	newBuildVars := make([]*entity.EnvVar, 0, len(req.BuildtimeEnvVars))
	newRuntimeVars := make([]*entity.EnvVar, 0, len(req.RuntimeEnvVars))
	for _, env := range req.SharedEnvVars {
		newRuntimeVars = append(newRuntimeVars, env.ToEntity(base.EnvVarKindShared))
	}
	for _, env := range req.RuntimeEnvVars {
		newRuntimeVars = append(newRuntimeVars, env.ToEntity(base.EnvVarKindRuntime))
	}
	for _, env := range req.BuildtimeEnvVars {
		newBuildVars = append(newBuildVars, env.ToEntity(base.EnvVarKindBuild))
	}

	envVars := &entity.EnvVars{
		Data: make([]*entity.EnvVar, 0, len(newRuntimeVars)+len(newBuildVars)),
	}
	envVars.Data = append(envVars.Data, newRuntimeVars...)
	envVars.Data = append(envVars.Data, newBuildVars...)
	setting.MustSetData(envVars)

	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Detect changes
	data.RuntimeVarsChange = !envvarhelper.Equal(data.CurrRuntimeVars, newRuntimeVars)
	data.BuildVarsChange = !envvarhelper.Equal(data.CurrBuildVars, newBuildVars)
}

func (uc *UC) buildAppEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateAppEnvVarsData,
) (runtimeData []*envvarservice.AppEnvVarData, err error) {
	if data.RuntimeVarsChange {
		// Validate the runtime env vars changes
		runtimeData, err = uc.envVarService.BuildEnvVarsForAllAppsInApp(ctx, db,
			&envvarservice.BuildEnvVarsInAppReq{
				App:          data.App,
				LoadOptions:  envvarservice.EnvLoadOptions{},
				BuildOptions: envvarservice.EnvBuildOptions{},
			}, true, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		var errors []string
		for _, appData := range runtimeData {
			errors = append(errors, appData.Errors()...)
		}
		if len(errors) > 0 {
			return nil, apperrors.Wrap(apperrors.ErrValidation).WithDisplayLevelHigh().
				WithExtraDetail(strings.Join(errors, "\n")) //nolint:govet
		}
	}

	if data.BuildVarsChange {
		// Validate the build-time env vars changes
		buildtimeData, err := uc.envVarService.BuildEnvVarsForAllAppsInApp(ctx, db,
			&envvarservice.BuildEnvVarsInAppReq{
				App: data.App,
				LoadOptions: envvarservice.EnvLoadOptions{
					BuildPhase: true,
				},
				BuildOptions: envvarservice.EnvBuildOptions{
					BuildPhaseOnly: true,
				},
			}, true, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		var errors []string
		for _, appData := range buildtimeData {
			errors = append(errors, appData.Errors()...)
		}
		if len(errors) > 0 {
			return nil, apperrors.Wrap(apperrors.ErrValidation).WithDisplayLevelHigh().
				WithExtraDetail(strings.Join(errors, "\n")) //nolint:govet
		}
	}

	return runtimeData, nil
}
