package notificationserviceimpl

import (
	"context"
	"errors"
	"time"

	goslack "github.com/slack-go/slack"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/slack"
)

const (
	slackSendRetryMax   = 2
	slackSendRetryDelay = 2 * time.Second
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

	for i := range slackSendRetryMax + 1 {
		if i > 0 {
			timer := time.NewTimer(slackSendRetryDelay * time.Duration(i))
			select {
			case <-ctx.Done():
				timer.Stop()
				return apperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}

		err = slack.NewClient().PostWebhook(ctx, webhookURL, "", msg)
		if err == nil {
			return nil
		}

		if !isRetryableSlackError(err) {
			return apperrors.Wrap(err)
		}
	}

	return apperrors.Wrap(err)
}

func isRetryableSlackError(err error) bool {
	if err == nil {
		return false
	}
	var rateLimitedErr *goslack.RateLimitedError
	if errors.As(err, &rateLimitedErr) {
		return true
	}
	var statusCodeErr *goslack.StatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.Retryable()
	}
	return true
}
