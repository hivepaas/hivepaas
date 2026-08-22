package notificationserviceimpl

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/discord"
)

const (
	discordSendRetryMax   = 2
	discordSendRetryDelay = 2 * time.Second
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

	for i := range discordSendRetryMax + 1 {
		if i > 0 {
			timer := time.NewTimer(discordSendRetryDelay * time.Duration(i))
			select {
			case <-ctx.Done():
				timer.Stop()
				return apperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}

		_, err = discord.NewClient().WebhookExecute(ctx, webhookURL, true, msg)
		if err == nil {
			return nil
		}

		if !isRetryableDiscordError(err) {
			return apperrors.Wrap(err)
		}
	}

	return apperrors.Wrap(err)
}

func isRetryableDiscordError(err error) bool {
	if err == nil {
		return false
	}
	var restErr *discordgo.RESTError
	if errors.As(err, &restErr) && restErr.Response != nil {
		statusCode := restErr.Response.StatusCode
		if statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError {
			return true
		}
		return false
	}
	return true
}
