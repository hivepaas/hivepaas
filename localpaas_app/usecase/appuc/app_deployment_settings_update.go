package appuc

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
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppDeploymentSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppDeploymentSettingsReq,
) (*appdto.UpdateAppDeploymentSettingsResp, error) {
	var persistingData *persistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppDeploymentSettingsData{}
		err := uc.loadAppDeploymentSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppDeploymentSettings(ctx, data, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.postTransactionAppDeploymentSettings(ctx, persistingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppDeploymentSettingsResp{}, nil
}

type updateAppDeploymentSettingsData struct {
	App                   *entity.App
	DeploymentSettings    *entity.Setting
	NewDeploymentSettings *entity.AppDeploymentSettings
	RegistryAuth          *entity.Setting
	Errors                []string // stores errors
	Warnings              []string // stores warnings
}

func (uc *UC) loadAppDeploymentSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppDeploymentSettingsReq,
	data *updateAppDeploymentSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.DeploymentSettings, _ = gofn.First(app.Settings)

	deploymentSettings := data.DeploymentSettings
	if deploymentSettings != nil && deploymentSettings.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newDeploymentSettings := req.ToEntity()
	data.NewDeploymentSettings = newDeploymentSettings

	// Make sure all reference settings used in this settings exist actively
	_, err = uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetSettingScope(),
		true, true, newDeploymentSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	repoSource := newDeploymentSettings.RepoSource
	if repoSource != nil {
		// Normalize repo type (currently supports git type only)
		if repoSource.RepoType == base.RepoTypeGit {
			repoSource.RepoRef = string(githelper.NormalizeRepoRef(repoSource.RepoRef))
		}

		// When the cluster has multiple nodes, the result image must be pushed to a registry
		// that can be accessed by all the nodes in the cluster.
		isMultiNode, err := uc.clusterService.IsMultiNode(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if isMultiNode && repoSource.PushToRegistry.ID == "" {
			return apperrors.Wrap(apperrors.ErrMultiNodeClusterRequireRegistryForImages)
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingAppDeploymentSettings(
	_ context.Context,
	data *updateAppDeploymentSettingsData,
	persistingData *persistingAppData,
) error {
	app := data.App
	setting := data.DeploymentSettings
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.SettingScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppDeployment,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppDeploymentSettingsVersion,
		}
		data.DeploymentSettings = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive
	setting.MustSetData(data.NewDeploymentSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Create a deployment and a task for it
	deployment, deploymentTask, err := uc.appService.CreateDeployment(app, data.NewDeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source: base.DeploymentTriggerSourceUser,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, deploymentTask)

	return nil
}

func (uc *UC) postTransactionAppDeploymentSettings(
	ctx context.Context,
	persistingData *persistingAppData,
) error {
	for _, task := range persistingData.UpsertingTasks {
		err := uc.taskQueue.ScheduleTask(ctx, task)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
