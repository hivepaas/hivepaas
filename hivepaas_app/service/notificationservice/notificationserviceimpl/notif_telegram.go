package notificationserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

const (
	telegramSendRetryMax   = 2
	telegramSendRetryDelay = 2 * time.Second
)

func (s *service) telegramSendMsg(
	ctx context.Context,
	setting *entity.IMTelegram,
	msg string,
) error {
	botToken, err := setting.BotToken.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}

	for i := range telegramSendRetryMax + 1 {
		if i > 0 {
			timer := time.NewTimer(telegramSendRetryDelay * time.Duration(i))
			select {
			case <-ctx.Done():
				timer.Stop()
				return apperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}

		err = telegram.NewClient().SendMessage(ctx, botToken, setting.ChatID, msg, "HTML")
		if err == nil {
			return nil
		}

		if !isRetryableTelegramError(err) {
			return apperrors.Wrap(err)
		}
	}

	return apperrors.Wrap(err)
}

func isRetryableTelegramError(err error) bool {
	if err == nil {
		return false
	}
	var statusCodeErr *telegram.StatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.Retryable()
	}
	return true
}
