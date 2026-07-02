package imserviceuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/hivepaas/hivepaas/services/im/discord"
	"github.com/hivepaas/hivepaas/services/im/slack"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

func (uc *UC) TestSendInstantMsg(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.TestSendInstantMsgReq,
) (_ *imservicedto.TestSendInstantMsgResp, err error) {
	switch req.Kind {
	case base.IMServiceKindSlack:
		err = slack.NewClient().PostWebhook(ctx, req.Slack.Webhook, "", req.TestMsg)
	case base.IMServiceKindDiscord:
		_, err = discord.NewClient().WebhookExecute(ctx, req.Discord.Webhook, true, req.TestMsg)
	case base.IMServiceKindTelegram:
		err = telegram.NewClient().SendMessage(ctx, req.Telegram.BotToken, req.Telegram.ChatID,
			req.TestMsg, "")
	}
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &imservicedto.TestSendInstantMsgResp{}, nil
}
