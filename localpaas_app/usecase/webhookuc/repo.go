package webhookuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
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
			return apperrors.New(err)
		}

		err = uc.processRepoWebhook(ctx, db, req, data, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		err = uc.appService.PersistAppData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
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
	Push      *repoPushEventData
	PRComment *repoPRCommentEventData
	PRClosed  *repoPRClosedEventData
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
		return apperrors.New(err)
	}
	_, err = setting.AsRepoWebhook()
	if err != nil {
		return apperrors.New(err)
	}
	data.WebhookSetting = setting
	return nil
}

func (uc *UC) processRepoWebhook(
	ctx context.Context,
	db database.Tx,
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
		return apperrors.New(apperrors.ErrWebhookTypeUnsupported).WithParam("Type", webhook.Kind)
	}
	if err != nil {
		return apperrors.New(err)
	}

	if eventData.Push != nil {
		if err := uc.processWebhookEventPush(ctx, db, eventData.Push, data, persistingData); err != nil {
			return apperrors.New(err)
		}
	}

	if eventData.PRComment != nil {
		if err := uc.processWebhookEventPRComment(ctx, db, eventData.PRComment, data, persistingData); err != nil {
			return apperrors.New(err)
		}
	}

	if eventData.PRClosed != nil {
		if err := uc.processWebhookEventPRClosed(ctx, db, eventData.PRClosed, data, persistingData); err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
