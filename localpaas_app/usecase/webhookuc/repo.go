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
	RepoWebhook *entity.RepoWebhook
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
	data *handleRepoWebhookData,
	persistingData *appservice.PersistingAppData,
) (err error) {
	eventData := &repoEventData{}
	switch data.RepoWebhook.Kind {
	case base.WebhookKindGithub:
		err = uc.parseGithubWebhook(req.Request, data.RepoWebhook.Secret, eventData)
	case base.WebhookKindGitlab:
		err = uc.parseGitlabWebhook(req.Request, data.RepoWebhook.Secret, eventData)
	case base.WebhookKindGitea:
		err = uc.parseGiteaWebhook(req.Request, data.RepoWebhook.Secret, eventData)
	case base.WebhookKindBitbucket:
		err = uc.parseBitbucketWebhook(req.Request, data.RepoWebhook.Secret, eventData)
	case base.WebhookKindGogs:
		err = uc.parseGogsWebhook(req.Request, data.RepoWebhook.Secret, eventData)
	default:
		return apperrors.NewUnsupported(apperrors.Fmt("Webhook kind '%v'", data.RepoWebhook.Kind))
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	if eventData.Push != nil {
		eventData.Push.parsedURL, err = vcsurl.Parse(eventData.Push.RepoURL)
		if err != nil {
			return apperrors.Wrap(err)
		}

		settings, err := uc.findAppDeploymentSettingsByPushEvent(ctx, db, eventData.Push)
		if err != nil {
			return apperrors.Wrap(err)
		}
		for _, setting := range settings {
			err = uc.createAppDeploymentByPushEvent(setting, eventData.Push, persistingData)
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
	data.RepoWebhook, err = setting.AsRepoWebhook()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) findAppDeploymentSettingsByPushEvent(
	ctx context.Context,
	db database.IDB,
	pushEvent *repoPushEventData,
) ([]*entity.Setting, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'activeMethod' = ?", base.DeploymentMethodRepo),
		bunex.SelectWhere("setting.data->>'repoRef' = ?", pushEvent.RepoRef),
		bunex.SelectWhere("setting.data->>'repoId' = ?", pushEvent.parsedURL.ID),

		bunex.SelectRelation("BelongToApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}

	validSettings := make([]*entity.Setting, 0, len(settings))
	for _, setting := range settings {
		app := setting.BelongToApp
		if app == nil || app.Status != base.AppStatusActive ||
			app.Project == nil || app.Project.Status != base.ProjectStatusActive {
			continue
		}
		shouldRedeploy, err := uc.shouldRedeployAppByPushEvent(ctx, db, app, pushEvent)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if shouldRedeploy {
			validSettings = append(validSettings, setting)
		}
	}

	return validSettings, nil
}

func (uc *UC) shouldRedeployAppByPushEvent(
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
		return false, apperrors.Wrap(err)
	}
	return len(deployments) == 0, nil
}

func (uc *UC) createAppDeploymentByPushEvent(
	setting *entity.Setting, // deployment setting
	pushEvent *repoPushEventData,
	persistingData *appservice.PersistingAppData,
) error {
	deploymentSettings, err := setting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}
	if deploymentSettings.RepoSource != nil && deploymentSettings.RepoSource.CommitHash != "" {
		deploymentSettings.RepoSource.CommitHash = ""
		setting.MustSetData(deploymentSettings)
		setting.UpdateVer++
		setting.UpdatedAt = timeutil.NowUTC()
		persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
	}

	app := setting.BelongToApp
	deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// Override target commit hash
	deployment.Settings.RepoSource.CommitHash = pushEvent.ChangeID
	// Set trigger for the deployment
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceRepoWebhook,
		ChangeID: pushEvent.ChangeID,
	}

	persistingData.UpsertingDeployments = append(persistingData.UpsertingDeployments, deployment)
	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
	return nil
}
