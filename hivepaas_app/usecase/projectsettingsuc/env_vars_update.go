package projectsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) UpdateProjectEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
) (*projectsettingsdto.UpdateProjectEnvVarsResp, error) {
	var buildData []*envvarservice.AppEnvVarData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareUpdatingProjectEnvVars(req, data, persistingData)

		if !data.BuildVarsChange && !data.RuntimeVarsChange {
			return nil
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Build and validate the env var changes
		buildData, err = uc.buildProjectEnvVars(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &projectsettingsdto.UpdateProjectEnvVarsResp{
		Meta: &basedto.Meta{},
	}

	// Apply the changes to every project envs and their apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, uc.db, buildData, true, true)
	if len(errMap) == 0 {
		return resp, nil
	}
	// NOTE: just show user a message instead of failing the request?
	resp.Meta.Warning = "Project updated successfully, but failed to apply changes to apps:"
	for i, e := range errMap {
		resp.Meta.Warning += fmt.Sprintf("\nApp '%v': ", buildData[i].App.Name) + e.Error()
	}

	return resp, nil
}

type updateProjectEnvVarsData struct {
	Project        *entity.Project
	EnvVarsSetting *entity.Setting

	CurrRuntimeVars []*entity.EnvVar
	CurrBuildVars   []*entity.EnvVar

	RuntimeVarsChange bool
	BuildVarsChange   bool
}

func (uc *UC) loadProjectEnvVarsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
	data *updateProjectEnvVarsData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE"), // NOTE: lock all loaded objects
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		),
		bunex.SelectRelation("ProjectEnvs.Apps"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project
	setting := project.GetSettingByType(base.SettingTypeEnvVar)
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

func (uc *UC) prepareUpdatingProjectEnvVars(
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
	data *updateProjectEnvVarsData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	setting := data.EnvVarsSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeProject,
			ObjectID:  project.ID,
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

func (uc *UC) buildProjectEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateProjectEnvVarsData,
) (runtimeData []*envvarservice.AppEnvVarData, err error) {
	if data.RuntimeVarsChange {
		// Validate the runtime env vars changes
		runtimeData, err = uc.envVarService.BuildEnvVarsForAllAppsInProject(ctx, db,
			&envvarservice.BuildEnvVarsInProjectReq{
				Project:      data.Project,
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
		buildTimeData, err := uc.envVarService.BuildEnvVarsForAllAppsInProject(ctx, db,
			&envvarservice.BuildEnvVarsInProjectReq{
				Project: data.Project,
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
