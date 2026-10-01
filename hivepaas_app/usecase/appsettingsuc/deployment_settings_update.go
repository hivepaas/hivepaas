package appsettingsuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppDeploymentSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppDeploymentSettingsReq,
) (*appsettingsdto.UpdateAppDeploymentSettingsResp, error) {
	var persistingData *persistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppDeploymentSettingsData{}
		err := uc.loadAppDeploymentSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppDeploymentSettings(ctx, auth, data, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "deployment", auditdetail.New().
			WithChangedFields(parseAppSetting(data.DeploymentSetting), data.NewDeploymentSettings))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, persistingData.UpsertingTasks...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var deploymentID, taskID string
	if deployment, _ := gofn.First(persistingData.UpsertingDeployments); deployment != nil {
		deploymentID = deployment.ID
	}
	if deploymentTask, _ := gofn.First(persistingData.UpsertingTasks); deploymentTask != nil {
		taskID = deploymentTask.ID
	}

	return &appsettingsdto.UpdateAppDeploymentSettingsResp{
		Data: &appsettingsdto.UpdateAppDeploymentSettingsDataResp{
			DeploymentID: deploymentID,
			TaskID:       taskID,
		},
	}, nil
}

type updateAppDeploymentSettingsData struct {
	App                   *entity.App
	DeploymentSetting     *entity.Setting
	NewDeploymentSettings *entity.AppDeploymentSettings
	RegistryAuthSetting   *entity.Setting
}

func (uc *UC) loadAppDeploymentSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppDeploymentSettingsReq,
	data *updateAppDeploymentSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppDeployment, base.SettingTypeAppKind),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app
	data.DeploymentSetting = app.GetSettingByType(base.SettingTypeAppDeployment)

	deploymentSettings := data.DeploymentSetting
	if deploymentSettings != nil && deploymentSettings.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	newDeploymentSettings, err := req.ToEntity()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.NewDeploymentSettings = newDeploymentSettings

	err = checkDeploymentOfKind(entity.IsFunctionKind(app.GetSettingByType(base.SettingTypeAppKind)),
		newDeploymentSettings)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Make sure all reference settings used in this settings exist actively
	refObjects := entity.NewRefObjects()
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, app.GetObjectScope(),
		true, newDeploymentSettings.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}

	// A build checks out its source and pushes its image: what it will need is
	// checked before the settings are saved.
	checkReq := &appdeploymentservice.CheckBuildSourceReq{RefObjects: refObjects}
	switch newDeploymentSettings.ActiveMethod {
	case base.DeploymentMethodRepo:
		checkReq.RepoSource = newDeploymentSettings.RepoSource
		checkReq.PushToRegistry = newDeploymentSettings.RepoSource.PushToRegistry
	case base.DeploymentMethodFunction:
		functionSource := newDeploymentSettings.FunctionSource
		if functionSource.Code.Repo != nil {
			checkReq.RepoSource = functionSource.Code.Repo.RepoSource()
		}
		checkReq.PushToRegistry = functionSource.PushToRegistry
	case base.DeploymentMethodImage:
		return nil
	}
	if err = uc.appDeploymentService.CheckBuildSource(ctx, checkReq); err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppDeploymentSettings(
	_ context.Context,
	auth *basedto.Auth,
	data *updateAppDeploymentSettingsData,
	persistingData *persistingAppData,
) error {
	app := data.App
	setting := data.DeploymentSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeApp,
			ObjectID:    app.ID,
			Type:        base.SettingTypeAppDeployment,
			Inheritable: true,
			CreatedAt:   timeNow,
			Version:     entity.CurrentAppDeploymentSettingsVersion,
		}
		data.DeploymentSetting = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive
	setting.MustSetData(data.NewDeploymentSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Create a deployment and a task for it
	deployment, deploymentTask, err := uc.appDeploymentService.CreateDeploymentAndTask(
		app, data.NewDeploymentSettings, appdeploymentservice.DeploymentArgs{})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceUser,
		SourceID: auth.User.ID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, deploymentTask)
	return nil
}
