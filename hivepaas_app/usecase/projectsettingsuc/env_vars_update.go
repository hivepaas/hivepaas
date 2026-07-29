package projectsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) UpdateProjectEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
) (*projectsettingsdto.UpdateProjectEnvVarsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareUpdatingProjectEnvVars(req, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// TODO: validate the env vars changes

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &projectsettingsdto.UpdateProjectEnvVarsResp{
		Meta: &basedto.Meta{},
	}

	// Apply the changes to every project envs and their apps
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectFor("UPDATE OF project"),
			bunex.SelectRelation("ProjectEnvs.Apps"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.projectService.ApplyEnvVarChangesToEnvs(ctx, db, &envvarservice.BuildEnvVarsInProjectReq{
			Project: project,
			BuildOptions: envvarservice.EnvBuildOptions{
				Sort: true,
			},
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		// NOTE: just show user a message instead of failing the request?
		resp.Meta.Warning = "Project updated successfully, but failed to apply changes to app: " + err.Error()
	}

	return resp, nil
}

type updateProjectEnvVarsData struct {
	Project        *entity.Project
	EnvVarsSetting *entity.Setting
}

func (uc *UC) loadProjectEnvVarsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
	data *updateProjectEnvVarsData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project
	data.EnvVarsSetting = project.GetSettingByType(base.SettingTypeEnvVar)

	if data.EnvVarsSetting != nil && data.EnvVarsSetting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
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

	envVars := &entity.EnvVars{
		Data: make([]*entity.EnvVar, 0, len(req.BuildtimeEnvVars)+len(req.RuntimeEnvVars)),
	}
	for _, env := range req.BuildtimeEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(base.EnvVarKindBuild))
	}
	for _, env := range req.RuntimeEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(base.EnvVarKindRuntime))
	}
	setting.MustSetData(envVars)

	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}
