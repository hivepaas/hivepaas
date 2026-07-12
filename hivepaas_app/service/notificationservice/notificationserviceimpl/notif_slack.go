package notificationserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/slack"
)

func (s *service) slackSendMsg(
	ctx context.Context,
	setting *entity.IMSlack,
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
