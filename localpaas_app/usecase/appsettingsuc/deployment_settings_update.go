package appsettingsuc

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
	"github.com/localpaas/localpaas/localpaas_app/pkg/gittool"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
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
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		err = uc.prepareUpdatingAppDeploymentSettings(ctx, auth, data, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.persistAppDeploymentSettingsRepoLinks(ctx, db, data)
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

	return &appsettingsdto.UpdateAppDeploymentSettingsResp{}, nil
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
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.DeploymentSetting = app.GetSettingByType(base.SettingTypeAppDeployment)

	deploymentSettings := data.DeploymentSetting
	if deploymentSettings != nil && deploymentSettings.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newDeploymentSettings, err := req.ToEntity()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.NewDeploymentSettings = newDeploymentSettings

	// Make sure all reference settings used in this settings exist actively
	refObjects, err := uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetObjectScope(),
		true, true, newDeploymentSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	if newDeploymentSettings.ActiveMethod == base.DeploymentMethodRepo {
		repoSource := newDeploymentSettings.RepoSource

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
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppDeployment,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppDeploymentSettingsVersion,
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
	deployment, deploymentTask, err := uc.appDeploymentService.CreateDeploymentAndTask(app, data.NewDeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
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

func (uc *UC) persistAppDeploymentSettingsRepoLinks(
	ctx context.Context,
	db database.Tx,
	data *updateAppDeploymentSettingsData,
) error {
	var repoIDs []string
	if data.NewDeploymentSettings.ActiveMethod == base.DeploymentMethodRepo {
		repoIDs = append(repoIDs, data.NewDeploymentSettings.RepoSource.RepoID)
	}
	err := uc.resLinkService.SetLinks(ctx, db, base.ResourceTypeApp, data.App.ID,
		base.ResourceTypeRepo, repoIDs)
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
