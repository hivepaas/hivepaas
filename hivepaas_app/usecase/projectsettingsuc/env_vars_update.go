package projectsettingsuc

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

		if !data.HasChanges {
			return nil
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Build and validate the env var changes
		buildData, err = uc.buildProjectEnvVars(ctx, db, data, true)
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
	Project           *entity.Project
	EnvVarsSetting    *entity.Setting
	CurrVars          []*entity.EnvVar
	HasChanges        bool
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

	data.EnvVarsSetting = setting
	if setting != nil {
		// Calculate current data to detect changes
		envVars, err := setting.AsEnvVars()
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.CurrVars = envVars.Data
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
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeProject,
			ObjectID:    project.ID,
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
		Data: make([]*entity.EnvVar, 0, len(req.RuntimeEnvVars)+len(req.BuildtimeEnvVars)),
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
	data.RuntimeVarsChange, _, data.BuildVarsChange = envvarhelper.CalcContentChanges(envVars.Data, data.CurrVars)
	// Detect any change even the order of the items
	data.HasChanges = data.RuntimeVarsChange || data.BuildVarsChange
	if !data.HasChanges {
		data.HasChanges = envvarhelper.Equal(envVars.Data, data.CurrVars)
	}
}

func (uc *UC) buildProjectEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateProjectEnvVarsData,
	inTx bool,
) (appEnvVarData []*envvarservice.AppEnvVarData, err error) {
	scope := data.Project.GetObjectScope()
	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	if data.RuntimeVarsChange {
		appEnvVarData, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope, false,
			nil, transaction, concurrency)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	if data.BuildVarsChange {
		// For build phase env vars, just validate them
		_, err = uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, scope, true,
			nil, transaction, concurrency)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return appEnvVarData, nil
}
