package imserviceuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/localpaas/localpaas/services/im/discord"
	"github.com/localpaas/localpaas/services/im/slack"
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
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imservicedto.TestSendInstantMsgResp{}, nil
}
