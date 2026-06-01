package appactionuc

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
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appactionuc/appactiondto"
)

func (uc *UC) DeployApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appactiondto.DeployAppReq,
) (*appactiondto.DeployAppResp, error) {
	var data *deployAppData
	var persistingData *persistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &deployAppData{}
		err := uc.loadAppActionSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppActionSettings(auth, req, data, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.persistAppData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.postTransactionAppActionSettings(ctx, persistingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	deployment, _ := gofn.First(persistingData.UpsertingDeployments)
	return &appactiondto.DeployAppResp{
		Data: &appactiondto.DeployAppDataResp{DeploymentID: deployment.ID},
	}, nil
}

type deployAppData struct {
	App                    *entity.App
	DeploymentSettings     *entity.Setting
	CurrDeploymentSettings *entity.AppDeploymentSettings
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) loadAppActionSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appactiondto.DeployAppReq,
	data *deployAppData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.DeploymentSettings, _ = gofn.First(app.Settings)

	if data.DeploymentSettings == nil || !data.DeploymentSettings.IsActive() {
		return apperrors.NewNotFound("AppActionSettings").
			WithMsgLog("app deployment settings not found")
	}

	// Parse the current deployment settings
	currSetting, err := data.DeploymentSettings.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to parse app deployment settings")
	}
	data.CurrDeploymentSettings = currSetting

	// Validate active deployment method
	req.ActiveMethod = gofn.Coalesce(req.ActiveMethod, currSetting.ActiveMethod)
	if req.ActiveMethod == "" {
		return apperrors.NewMissing("Deployment method")
	}

	// Normalize repo ref
	if req.RepoSource != nil && req.RepoSource.RepoRef != "" && currSetting.RepoSource != nil {
		if currSetting.RepoSource.RepoType == base.RepoTypeGit {
			req.RepoSource.RepoRef = string(githelper.NormalizeRepoRef(req.RepoSource.RepoRef))
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingAppActionSettings(
	auth *basedto.Auth,
	req *appactiondto.DeployAppReq,
	data *deployAppData,
	persistingData *persistingAppData,
) error {
	app := data.App
	setting := data.DeploymentSettings
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive

	currDeploymentSettings := data.CurrDeploymentSettings
	err := req.ApplyTo(currDeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}

	setting.MustSetData(currDeploymentSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Create a deployment and a task for it
	deployment, deploymentTask, err := uc.appDeploymentService.CreateDeploymentAndTask(app, currDeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceAPI,
		ID:       auth.User.ID,
		ChangeID: req.ChangeID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, deploymentTask)

	return nil
}

func (uc *UC) persistAppData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) postTransactionAppActionSettings(
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
