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
		data := &handleRepoWebhookData{}
		persistingData = &appservice.PersistingAppData{}

		err := uc.loadWebhookSettings(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.processRepoWebhook(ctx, db, req, data, persistingData)
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

type handleRepoWebhookData struct {
	WebhookSetting *entity.Setting
}

type repoEventData struct {
	Push *repoPushEventData
}

type repoPushEventData struct {
	RepoRef  string
	RepoURL  string
	RepoID   string
	ChangeID string
}

func (uc *UC) processRepoWebhook(
	ctx context.Context,
	db database.IDB,
	req *webhookdto.HandleRepoWebhookReq,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) (err error) {
	webhook := data.WebhookSetting.MustAsRepoWebhook()
	eventData := &repoEventData{}
	switch webhook.Kind {
	case base.WebhookKindGithub:
		err = uc.parseGithubWebhook(req.Request, webhook.Secret, eventData)
	case base.WebhookKindGitlab:
		err = uc.parseGitlabWebhook(req.Request, webhook.Secret, eventData)
	case base.WebhookKindGitea:
		err = uc.parseGiteaWebhook(req.Request, webhook.Secret, eventData)
	case base.WebhookKindBitbucket:
		err = uc.parseBitbucketWebhook(req.Request, webhook.Secret, eventData)
	case base.WebhookKindGogs:
		err = uc.parseGogsWebhook(req.Request, webhook.Secret, eventData)
	default:
		return apperrors.NewUnsupported(apperrors.Fmt("Webhook kind '%v'", webhook.Kind))
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	if eventData.Push != nil {
		parsedURL, err := vcsurl.Parse(eventData.Push.RepoURL)
		if err != nil {
			return apperrors.Wrap(err)
		}
		eventData.Push.RepoID = parsedURL.ID

		apps, err := uc.findAppsToRedeployByPushEvent(ctx, db, eventData.Push)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, app := range apps {
			err = uc.createAppDeploymentByPushEvent(app, eventData.Push, data, persistingData)
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
	data *handleRepoWebhookData,
) error {
	setting, err := uc.settingRepo.GetByID(ctx, db, nil, "", req.ID, true,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeRepoWebhook, base.SettingTypeGithubApp),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, err = setting.AsRepoWebhook()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.WebhookSetting = setting
	return nil
}

func (uc *UC) findAppsToRedeployByPushEvent(
	ctx context.Context,
	db database.IDB,
	pushEvent *repoPushEventData,
) ([]*entity.App, error) {
	apps, _, err := uc.appRepo.List(ctx, db, "", nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			bunex.SelectWhere("project.status = ?", base.ProjectStatusActive),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),

		bunex.SelectJoin("JOIN res_links ON res_links.src_id = app.id"),
		bunex.SelectWhere("res_links.deleted_at IS NULL"),
		bunex.SelectWhere("res_links.dst_type = ?", base.ResourceTypeRepo),
		bunex.SelectWhere("res_links.dst_id = ?", pushEvent.RepoID),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, nil
	}

	matchingApps := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		if app.Project == nil || app.Project.Status != base.ProjectStatusActive {
			continue
		}
		shouldRedeploy, err := uc.shouldRedeployAppByPushEvent(ctx, db, app, pushEvent)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if shouldRedeploy {
			matchingApps = append(matchingApps, app)
		}
	}
	return matchingApps, nil
}

func (uc *UC) shouldRedeployAppByPushEvent(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	pushEvent *repoPushEventData,
) (bool, error) {
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return false, nil
	}
	deploymentSettings := deploymentSetting.MustAsAppDeploymentSettings()
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo {
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
		bunex.SelectWhere("deployment.trigger->>'changeId' = ?", pushEvent.ChangeID),
	)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	return len(deployments) == 0, nil
}

func (uc *UC) createAppDeploymentByPushEvent(
	app *entity.App,
	pushEvent *repoPushEventData,
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) error {
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
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
