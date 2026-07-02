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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) UpdateProjectEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.UpdateProjectEnvVarsReq,
) (*projectsettingsdto.UpdateProjectEnvVarsResp, error) {
	var data *updateProjectEnvVarsData
	var persistingData *persistingProjectData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData = &persistingProjectData{}
		uc.prepareUpdatingProjectEnvVars(req, data, persistingData)

		// TODO: Do we need to re-apply the ENVs to the apps?

		// TODO: how to make sure the changes not break apps
		// if they use any of ENVs within the project.

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &projectsettingsdto.UpdateProjectEnvVarsResp{}, nil
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
		return apperrors.New(err)
	}
	data.Project = project
	data.EnvVarsSetting = project.GetSettingByType(base.SettingTypeEnvVar)

	if data.EnvVarsSetting != nil && data.EnvVarsSetting.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
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
		envVars.Data = append(envVars.Data, env.ToEntity(true))
	}
	for _, env := range req.RuntimeEnvVars {
		envVars.Data = append(envVars.Data, env.ToEntity(false))
	}
	setting.MustSetData(envVars)

	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}
