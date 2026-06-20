package appactionuc

import (
	"context"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/gittool"
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
		err := uc.loadAppDeploymentSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppDeploymentSettings(auth, req, data, persistingData)
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

	err = uc.postTransactionAppDeploymentSettings(ctx, persistingData)
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
	DeploymentSettingEnt   *entity.Setting
	CurrDeploymentSettings *entity.AppDeploymentSettings
	NewDeploymentSettings  *entity.AppDeploymentSettings
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) loadAppDeploymentSettingsForUpdate(
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
	data.DeploymentSettingEnt, _ = gofn.First(app.Settings)

	if data.DeploymentSettingEnt == nil || !data.DeploymentSettingEnt.IsActive() {
		return apperrors.NewNotFound("App deployment settings").
			WithMsgLog("app deployment settings not found")
	}

	// Parse the current deployment settings
	currSettings, err := data.DeploymentSettingEnt.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to parse app deployment settings")
	}
	data.CurrDeploymentSettings = currSettings

	newSettings, err := copier.CopyAs(currSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.NewDeploymentSettings = newSettings
	if err = req.ApplyTo(newSettings); err != nil {
		return apperrors.Wrap(err)
	}

	// Make sure all reference settings used in this settings exist actively
	refObjects, err := uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetObjectScope(),
		true, true, newSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Validate active deployment method
	if newSettings.ActiveMethod == "" {
		return apperrors.NewMissing("Deployment method")
	}

	switch newSettings.ActiveMethod {
	case base.DeploymentMethodImage:
		// Do nothing

	case base.DeploymentMethodRepo:
		repoSource := newSettings.RepoSource

		// When the cluster has multiple nodes, the result image must be pushed to a registry
		// that can be accessed by all the nodes in the cluster.
		isMultiNode, err := uc.clusterService.IsMultiNode(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if isMultiNode && repoSource.PushToRegistry.ID == "" {
			return apperrors.Wrap(apperrors.ErrMultiNodeClusterRequireRegistryForImages)
		}

		// Validate existence of repo and ref
		switch repoSource.RepoType { //nolint:gocritic
		case base.RepoTypeGit:
			// TODO: do not check commit hash for now, that's so slow
			err := gittool.ValidateWithGitCli(ctx, &gittool.ValidationOptions{
				URL:           repoSource.RepoURL,
				Credentials:   refObjects.RefSettings[repoSource.Credentials.ID],
				ReferenceName: plumbing.ReferenceName(repoSource.RepoRef),
			})
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

	default:
		return apperrors.NewValueInvalid("deployment method")
	}

	return nil
}

func (uc *UC) prepareUpdatingAppDeploymentSettings(
	auth *basedto.Auth,
	req *appactiondto.DeployAppReq,
	data *deployAppData,
	persistingData *persistingAppData,
) error {
	app := data.App
	setting := data.DeploymentSettingEnt
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	setting.ExpireAt = time.Time{}
	setting.Status = base.SettingStatusActive

	setting.MustSetData(data.NewDeploymentSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)

	// Create a deployment and a task for it
	deployment, deploymentTask, err := uc.appDeploymentService.CreateDeploymentAndTask(app, data.NewDeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// Set NoCache for the current deployment only if configured
	deployment.Settings.NoCache = req.NoCache
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceAPI,
		SourceID: auth.User.ID,
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
