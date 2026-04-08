package projectuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProjectEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectEnvVarsReq,
) (*projectdto.UpdateProjectEnvVarsResp, error) {
	var data *updateProjectEnvVarsData
	var persistingData *persistingProjectData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData = &persistingProjectData{}
		uc.prepareUpdatingProjectEnvVars(req, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectEnvVarsResp{}, nil
}

type updateProjectEnvVarsData struct {
	Project  *entity.Project
	EnvVars  *entity.Setting
	Errors   []string // stores errors
	Warnings []string // stores warnings
}

func (uc *UC) loadProjectEnvVarsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *projectdto.UpdateProjectEnvVarsReq,
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
	data.EnvVars, _ = gofn.First(project.Settings)

	if data.EnvVars != nil && data.EnvVars.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingProjectEnvVars(
	req *projectdto.UpdateProjectEnvVarsReq,
	data *updateProjectEnvVarsData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	setting := data.EnvVars
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.SettingScopeProject,
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
