package webhookuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc/webhookdto"
)

func (uc *UC) HandleRepoWebhook(
	ctx context.Context,
	req *webhookdto.HandleRepoWebhookReq,
) (*webhookdto.HandleRepoWebhookResp, error) {
	db := uc.db

	data := &handleRepoWebhookData{}
	if err := uc.loadWebhookSettings(ctx, db, req, data); err != nil {
		return nil, apperrors.New(err)
	}

	if err := uc.processRepoWebhook(ctx, db, req, data); err != nil {
		return nil, apperrors.New(err)
	}

	return &webhookdto.HandleRepoWebhookResp{}, nil
}

type handleRepoWebhookData struct {
	WebhookSetting *entity.Setting
}

type repoEventData struct {
	Push           *repoPushEventData
	PRComment      *repoPRCommentEventData
	PRClosed       *repoPRClosedEventData
	PRSynchronized *repoPRSynchronizedEventData
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
	db database.IDB,
	req *webhookdto.HandleRepoWebhookReq,
	data *handleRepoWebhookData,
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

	switch {
	case eventData.Push != nil:
		err = uc.processWebhookEventPush(ctx, db, eventData.Push, data)
	case eventData.PRComment != nil:
		err = uc.processWebhookEventPRComment(ctx, db, eventData.PRComment, data)
	case eventData.PRClosed != nil:
		err = uc.processWebhookEventPRClosed(ctx, db, eventData.PRClosed, data)
	case eventData.PRSynchronized != nil:
		err = uc.processWebhookEventPRSynchronized(ctx, db, eventData.PRSynchronized, data)
	}
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
