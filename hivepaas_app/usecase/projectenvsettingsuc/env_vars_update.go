package projectenvsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

func (uc *UC) UpdateProjectEnvEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvsettingsdto.UpdateProjectEnvEnvVarsReq,
) (*projectenvsettingsdto.UpdateProjectEnvEnvVarsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateProjectEnvVarsData{}
		err := uc.loadProjectEnvEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectEnvData{}
		uc.prepareUpdatingProjectEnvEnvVars(req, data, persistingData)

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

	resp := &projectenvsettingsdto.UpdateProjectEnvEnvVarsResp{
		Meta: &basedto.Meta{},
	}

	// Apply the changes to every belonging apps
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectEnv, err := uc.projectEnvRepo.GetByID(ctx, db, req.ProjectID, req.ProjectEnvID,
			bunex.SelectFor("UPDATE OF project_env"),
			bunex.SelectRelation("Apps",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.projectService.ApplyEnvVarChangesToApps(ctx, db, &envvarservice.BuildEnvVarsInProjectEnvReq{
			ProjectEnv: projectEnv,
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
		resp.Meta.Warning = "Project env updated successfully, but failed to apply changes to app: " + err.Error()
	}

	return resp, nil
}

type updateProjectEnvVarsData struct {
	ProjectEnv     *entity.ProjectEnv
	EnvVarsSetting *entity.Setting
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
	data.EnvVarsSetting = projectEnv.GetSettingByType(base.SettingTypeEnvVar)

	if data.EnvVarsSetting != nil && data.EnvVarsSetting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
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
