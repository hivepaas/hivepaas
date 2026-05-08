package notificationserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/services/im/slack"
)

func (s *service) slackSendMsg(
	ctx context.Context,
	setting *entity.Slack,
	msg string,
) error {
	webhookURL, err := setting.Webhook.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = slack.NewClient().PostWebhook(ctx, webhookURL, "", msg)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
