package notificationserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/discord"
)

func (s *service) discordSendMsg(
	ctx context.Context,
	setting *entity.IMDiscord,
	msg string,
) error {
	webhookURL, err := setting.Webhook.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, err = discord.NewClient().WebhookExecute(ctx, webhookURL, true, msg)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
