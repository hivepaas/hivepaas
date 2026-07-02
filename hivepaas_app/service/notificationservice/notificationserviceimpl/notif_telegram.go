package notificationserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

func (s *service) telegramSendMsg(
	ctx context.Context,
	setting *entity.IMTelegram,
	msg string,
) error {
	botToken, err := setting.BotToken.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}
	err = telegram.NewClient().SendMessage(ctx, botToken, setting.ChatID, msg, "HTML")
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
