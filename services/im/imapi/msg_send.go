package imapi

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/im/discord"
	"github.com/hivepaas/hivepaas/services/im/slack"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

func SendMessage(
	ctx context.Context,
	setting *entity.Setting,
	msg string,
) error {
	if msg == "" {
		return nil
	}
	if setting == nil {
		return apperrors.NewMissing("IM service setting")
	}
	if setting.Type != base.SettingTypeIMService {
		return apperrors.Wrap(apperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}

	imService, err := setting.AsIMService()
	if err != nil {
		return apperrors.Wrap(err)
	}
	if imService == nil {
		return apperrors.NewMissing("IM service setting")
	}

	switch base.IMServiceKind(setting.Kind) {
	case base.IMServiceKindSlack:
		if imService.Slack == nil {
			return apperrors.NewMissing("Slack setting")
		}
		webhookURL, err := imService.Slack.Webhook.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = slack.NewClient().PostWebhook(ctx, webhookURL, "", msg)
		if err != nil {
			return apperrors.Wrap(err)
		}

	case base.IMServiceKindDiscord:
		if imService.Discord == nil {
			return apperrors.NewMissing("Discord setting")
		}
		webhookURL, err := imService.Discord.Webhook.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
		_, err = discord.NewClient().WebhookExecute(ctx, webhookURL, true, msg)
		if err != nil {
			return apperrors.Wrap(err)
		}

	case base.IMServiceKindTelegram:
		if imService.Telegram == nil {
			return apperrors.NewMissing("Telegram setting")
		}
		botToken, err := imService.Telegram.BotToken.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = telegram.NewClient().SendMessage(ctx, botToken, imService.Telegram.ChatID, msg, "HTML")
		if err != nil {
			return apperrors.Wrap(err)
		}

	default:
		return apperrors.Wrap(apperrors.ErrIMServiceUnsupported).WithParam("Name", setting.Kind)
	}

	return nil
}

func SendMessageWithRetry(
	ctx context.Context,
	setting *entity.Setting,
	msg string,
	retryMax int,
	retryDelay time.Duration,
) error {
	if msg == "" {
		return nil
	}
	if setting == nil {
		return apperrors.NewMissing("IM service setting")
	}
	return retryExecute(ctx, func() error {
		return SendMessage(ctx, setting, msg)
	}, base.IMServiceKind(setting.Kind), retryMax, retryDelay)
}
