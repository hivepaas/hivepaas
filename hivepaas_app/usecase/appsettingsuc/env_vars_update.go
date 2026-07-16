package appsettingsuc

import (
	"context"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppEnvVarsReq,
) (*appsettingsdto.UpdateAppEnvVarsResp, error) {
	var data *updateAppEnvVarsData
	var persistingData *persistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateAppEnvVarsData{}
		err := uc.loadAppEnvVarsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		uc.prepareUpdatingAppEnvVars(req, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppEnvVars(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppEnvVarsResp{}, nil
}

type updateAppEnvVarsData struct {
	App            *entity.App
	EnvVarsSetting *entity.Setting
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
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.EnvVarsSetting = app.GetSettingByType(base.SettingTypeEnvVar)

	if data.EnvVarsSetting != nil && data.EnvVarsSetting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
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

func (uc *UC) applyAppEnvVars(
	ctx context.Context,
	db database.Tx,
	data *updateAppEnvVarsData,
) error {
	app := data.App
	envs, _, err := uc.envVarService.BuildAppEnvVars(ctx, db, app, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	envVars := make([]string, 0, len(envs))
	var errors []string
	for _, env := range envs {
		envVars = append(envVars, env.ToString("="))
		errors = append(errors, env.Errors...)
	}

	if len(errors) > 0 {
		return apperrors.Wrap(apperrors.ErrEnvVarContainInvalidReference).WithDisplayLevelHigh().
			WithExtraDetail("%s", strings.Join(errors, "\n"))
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if service.Spec.TaskTemplate.ContainerSpec == nil {
		service.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	service.Spec.TaskTemplate.ContainerSpec.Env = envVars

	_, err = uc.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
