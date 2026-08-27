package imapi

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/im/discord"
	"github.com/hivepaas/hivepaas/services/im/lark"
	"github.com/hivepaas/hivepaas/services/im/slack"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

//nolint:gocognit
func SendMessage(
	ctx context.Context,
	setting *entity.Setting,
	msg string,
) error {
	if msg == "" {
		return nil
	}
	if setting == nil {
		return hperrors.NewMissing("IM service setting")
	}
	if setting.Type != base.SettingTypeIMService {
		return hperrors.Wrap(hperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}

	imService, err := setting.AsIMService()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if imService == nil {
		return hperrors.NewMissing("IM service setting")
	}

	switch base.IMServiceKind(setting.Kind) {
	case base.IMServiceKindSlack:
		if imService.Slack == nil {
			return hperrors.NewMissing("Slack setting")
		}
		webhookURL, err := imService.Slack.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = slack.NewClient().PostWebhook(ctx, webhookURL, "", msg)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.IMServiceKindDiscord:
		if imService.Discord == nil {
			return hperrors.NewMissing("Discord setting")
		}
		webhookURL, err := imService.Discord.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		_, err = discord.NewClient().WebhookExecute(ctx, webhookURL, true, msg)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.IMServiceKindTelegram:
		if imService.Telegram == nil {
			return hperrors.NewMissing("Telegram setting")
		}
		botToken, err := imService.Telegram.BotToken.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = telegram.NewClient().SendMessage(ctx, botToken, imService.Telegram.ChatID, msg, "HTML")
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.IMServiceKindLark:
		if imService.Lark == nil {
			return hperrors.NewMissing("Lark setting")
		}
		webhookURL, err := imService.Lark.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		secret, err := imService.Lark.Secret.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = lark.NewClient().PostWebhook(ctx, webhookURL, secret, msg)
		if err != nil {
			return hperrors.Wrap(err)
		}

	default:
		return hperrors.Wrap(hperrors.ErrIMServiceUnsupported).WithParam("Name", setting.Kind)
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
		return hperrors.NewMissing("IM service setting")
	}
	return retryExecute(ctx, func() error {
		return SendMessage(ctx, setting, msg)
	}, base.IMServiceKind(setting.Kind), retryMax, retryDelay)
}
