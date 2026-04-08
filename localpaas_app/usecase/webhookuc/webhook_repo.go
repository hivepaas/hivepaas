package webhookuc

import (
	"context"
	"time"

	"github.com/gitsight/go-vcsurl"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) HandleRepoWebhook(
	ctx context.Context,
	req *webhookdto.HandleRepoWebhookReq,
) (*webhookdto.HandleRepoWebhookResp, error) {
	var persistingData *appservice.PersistingAppData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		persistingData = &appservice.PersistingAppData{}

		webhookSettings, err := uc.loadWebhookSettings(ctx, db, req)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(webhookSettings) == 0 { // no matching webhook setting found
			return nil
		}

		err = uc.processRepoWebhook(ctx, db, req, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.appService.PersistAppData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Schedule deployment tasks
	for _, task := range persistingData.UpsertingTasks {
		_ = uc.taskQueue.ScheduleTask(ctx, task)
	}

	return &webhookdto.HandleRepoWebhookResp{}, nil
}

type repoEventData struct {
	Push *repoPushEventData
}

type repoPushEventData struct {
	RepoRef  string
	RepoURL  string
	ChangeID string

	parsedURL *vcsurl.VCS
}

func (uc *UC) processRepoWebhook(
	ctx context.Context,
	db database.IDB,
	req *webhookdto.HandleRepoWebhookReq,
	persistingData *appservice.PersistingAppData,
) (err error) {
	data := &repoEventData{}
	switch req.WebhookKind {
	case base.WebhookKindGithub:
		err = uc.processGithubWebhook(req, data)
	case base.WebhookKindGitlab:
		err = uc.processGitlabWebhook(req, data)
	case base.WebhookKindGitea:
		err = uc.processGiteaWebhook(req, data)
	case base.WebhookKindBitbucket:
		err = uc.processBitbucketWebhook(req, data)
	case base.WebhookKindGogs:
		err = uc.processGogsWebhook(req, data)
	case base.WebhookKindAzureDevOps:
		err = uc.processAzureDevOpsWebhook(req, data)
	default:
		return apperrors.New(apperrors.ErrUnsupported).
			WithMsgLog("webhook kind '%s' not supported", req.WebhookKind)
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.Push != nil {
		apps, err := uc.findAppsToRedeployByPushEvent(ctx, db, data.Push)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, app := range apps {
			err = uc.createAppDeploymentByPushEvent(app, data.Push, persistingData)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}
	return nil
}

func (uc *UC) loadWebhookSettings(
	ctx context.Context,
	db database.IDB,
	req *webhookdto.HandleRepoWebhookReq,
) ([]*entity.Setting, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.type = ?", base.SettingTypeRepoWebhook),
			bunex.SelectWhere("setting.data->>'secret' = ?", req.Secret),
		),
		bunex.SelectWhereOrGroup(
			bunex.SelectWhere("setting.type = ?", base.SettingTypeGithubApp),
			bunex.SelectWhere("setting.data->>'webhook_secret' = ?", req.Secret),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return settings, nil
}

func (uc *UC) findAppsToRedeployByPushEvent(
	ctx context.Context,
	db database.IDB,
	pushEvent *repoPushEventData,
) ([]*entity.App, error) {
	apps, _, err := uc.appRepo.List(ctx, db, "", nil,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectJoin("JOIN projects ON projects.id = app.project_id"),
		bunex.SelectWhere("projects.status = ?", base.ProjectStatusActive),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, nil
	}

	pushEvent.parsedURL, err = vcsurl.Parse(pushEvent.RepoURL)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	appsToRedeploy := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		matching, err := uc.shouldRedeployAppByPushEvent(ctx, db, app, pushEvent)
		if err == nil && matching {
			appsToRedeploy = append(appsToRedeploy, app)
		}
	}

	return appsToRedeploy, nil
}

func (uc *UC) shouldRedeployAppByPushEvent(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	pushEvent *repoPushEventData,
) (bool, error) {
	if len(app.Settings) == 0 {
		return false, nil
	}
	deploymentSettings, err := app.Settings[0].AsAppDeploymentSettings()
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	shouldRedeploy := false
	for {
		repoSource := deploymentSettings.RepoSource
		if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo || repoSource == nil {
			break
		}
		if pushEvent.RepoRef != repoSource.RepoRef {
			break
		}
		if pushEvent.RepoURL == repoSource.RepoURL {
			shouldRedeploy = true
			break
		}
		parsedURL, err := vcsurl.Parse(repoSource.RepoURL)
		if err != nil {
			return false, apperrors.Wrap(err)
		}
		shouldRedeploy = parsedURL.ID == pushEvent.parsedURL.ID
		break //nolint:staticcheck
	}

	if !shouldRedeploy {
		return false, nil
	}

	// Make sure there is no duplicated deployment having the same `change id`
	if pushEvent.ChangeID == "" {
		return true, nil
	}
	deployments, _, err := uc.deploymentRepo.List(ctx, db, app.ID, nil,
		bunex.SelectColumns("id"),
		bunex.SelectLimit(1),
		bunex.SelectWhere("deployment.created_at > ?", timeutil.NowUTC().Add(-time.Minute)),
		bunex.SelectWhere("deployment.trigger->>'source' = ?", base.DeploymentTriggerSourceRepoWebhook),
		bunex.SelectWhere("deployment.trigger->>'id' = ?", pushEvent.ChangeID),
	)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	return len(deployments) == 0, nil
}

func (uc *UC) createAppDeploymentByPushEvent(
	app *entity.App,
	pushEvent *repoPushEventData,
	persistingData *appservice.PersistingAppData,
) error {
	deploymentSettings, err := app.Settings[0].AsAppDeploymentSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}

	deployment, task, err := uc.appService.CreateDeployment(app, deploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source: base.DeploymentTriggerSourceRepoWebhook,
		ID:     pushEvent.ChangeID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
	return nil
}
