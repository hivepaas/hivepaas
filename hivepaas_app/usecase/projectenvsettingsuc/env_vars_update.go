package projectenvsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

func (uc *UC) UpdateProjectEnvEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvsettingsdto.UpdateProjectEnvEnvVarsReq,
) (*projectenvsettingsdto.UpdateProjectEnvEnvVarsResp, error) {
	var buildData []*envvarservice.AppEnvVarData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectEnvData{}
		uc.prepareUpdatingProjectEnvEnvVars(req, data, persistingData)

		if !data.BuildVarsChange && !data.RuntimeVarsChange {
			return nil
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Build and validate the env var changes
		buildData, err = uc.buildProjectEnvEnvVars(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &projectenvsettingsdto.UpdateProjectEnvEnvVarsResp{
		Meta: &basedto.Meta{},
	}

	// Apply the changes to every belonging apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, uc.db, buildData, true, true)
	if len(errMap) == 0 {
		return resp, nil
	}
	// NOTE: just show user a message instead of failing the request?
	resp.Meta.Warning = "Project environment updated successfully, but failed to apply changes to apps:"
	for i, e := range errMap {
		resp.Meta.Warning += fmt.Sprintf("\nApp '%v': ", buildData[i].App.Name) + e.Error()
	}

	return resp, nil
}

type updateProjectEnvVarsData struct {
	ProjectEnv     *entity.ProjectEnv
	EnvVarsSetting *entity.Setting

	CurrRuntimeVars []*entity.EnvVar
	CurrBuildVars   []*entity.EnvVar

	RuntimeVarsChange bool
	BuildVarsChange   bool
}

func (uc *UC) loadProjectEnvEnvVarsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *projectenvsettingsdto.UpdateProjectEnvEnvVarsReq,
	data *updateProjectEnvVarsData,
) error {
	projectEnv, err := uc.projectEnvRepo.GetByID(ctx, db, req.ProjectID, req.ProjectEnvID,
		bunex.SelectFor("UPDATE OF project_env"),
		bunex.SelectRelation("Apps",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.ProjectEnv = projectEnv
	setting := projectEnv.GetSettingByType(base.SettingTypeEnvVar)
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

func (uc *UC) prepareUpdatingProjectEnvEnvVars(
	req *projectenvsettingsdto.UpdateProjectEnvEnvVarsReq,
	data *updateProjectEnvVarsData,
	persistingData *persistingProjectEnvData,
) {
	projectEnv := data.ProjectEnv
	setting := data.EnvVarsSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeProjectEnv,
			ObjectID:  projectEnv.ID,
			Type:      base.SettingTypeEnvVar,
			CreatedAt: timeNow,
			Version:   entity.CurrentEnvVarsVersion,
		}
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive

	newRuntimeVars := make([]*entity.EnvVar, 0, len(req.RuntimeEnvVars))
	newBuildVars := make([]*entity.EnvVar, 0, len(req.BuildtimeEnvVars))
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

func (uc *UC) buildProjectEnvEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateProjectEnvVarsData,
) (runtimeData []*envvarservice.AppEnvVarData, err error) {
	if data.RuntimeVarsChange {
		// Validate the runtime env vars changes
		runtimeData, err = uc.envVarService.BuildEnvVarsForAllAppsInProjectEnv(ctx, db,
			&envvarservice.BuildEnvVarsInProjectEnvReq{
				ProjectEnv:   data.ProjectEnv,
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
				WithExtraDetail("%s", strings.Join(errors, "\n"))
		}
	}

	if data.BuildVarsChange {
		// Validate the build-time env vars changes
		buildTimeData, err := uc.envVarService.BuildEnvVarsForAllAppsInProjectEnv(ctx, db,
			&envvarservice.BuildEnvVarsInProjectEnvReq{
				ProjectEnv: data.ProjectEnv,
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
		for _, appData := range buildTimeData {
			errors = append(errors, appData.Errors()...)
		}
		if len(errors) > 0 {
			return nil, apperrors.Wrap(apperrors.ErrValidation).WithDisplayLevelHigh().
				WithExtraDetail("%s", strings.Join(errors, "\n"))
		}
	}

	return runtimeData, nil
}
