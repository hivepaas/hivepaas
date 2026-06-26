package webhookuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
)

type repoPushEventData struct {
	RepoRef  string
	RepoURL  string
	RepoID   string
	ChangeID string
}

func (uc *UC) shouldCreateAppDeploymentByPushEvent(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	pushEvent *repoPushEventData,
) (bool, error) {
	// Make sure there is no duplicated deployment having the same `change id`
	if pushEvent.ChangeID == "" {
		return true, nil
	}
	deployments, _, err := uc.deploymentRepo.List(ctx, db, app.ID, nil,
		bunex.SelectColumns("id"),
		bunex.SelectLimit(1),
		bunex.SelectWhere("deployment.created_at > ?", timeutil.NowUTC().Add(-time.Minute)),
		bunex.SelectWhere("deployment.trigger->>'source' = ?", base.DeploymentTriggerSourceRepoWebhook),
		bunex.SelectWhere("deployment.trigger->>'changeId' = ?", pushEvent.ChangeID),
	)
	if err != nil {
		return false, apperrors.New(err)
	}
	return len(deployments) == 0, nil
}

func (uc *UC) createAppDeploymentByPushEvent(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	pushEvent *repoPushEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	shouldRedeploy, err := uc.shouldCreateAppDeploymentByPushEvent(ctx, db, app, pushEvent)
	if err != nil {
		return apperrors.New(err)
	}
	if !shouldRedeploy {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.New(err)
	}
	if deploymentSettings.RepoSource != nil && deploymentSettings.RepoSource.CommitHash != "" {
		deploymentSettings.RepoSource.CommitHash = ""
		deploymentSetting.MustSetData(deploymentSettings)
		deploymentSetting.UpdateVer++
		deploymentSetting.UpdatedAt = timeutil.NowUTC()
		persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, deploymentSetting)
	}

	deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings)
	if err != nil {
		return apperrors.New(err)
	}
	// Override target commit hash
	deployment.Settings.RepoSource.CommitHash = pushEvent.ChangeID
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceRepoWebhook,
		SourceID: data.WebhookSetting.ID,
		ChangeID: pushEvent.ChangeID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
	return nil
}
