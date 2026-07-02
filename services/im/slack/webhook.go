package slack

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/slack-go/slack"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

type WebhookMessageOption func(*slack.WebhookMessage)

func (c *Client) PostWebhook(ctx context.Context, webhookURL, channel, text string,
	options ...WebhookMessageOption) error {
	msg := &slack.WebhookMessage{}
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		if err := json.Unmarshal(reflectutil.UnsafeStrToBytes(text), msg); err != nil {
			msg.Text = text
		}
	} else {
		msg.Text = text
	}
	if channel != "" {
		msg.Channel = channel
	}

	for _, opt := range options {
		opt(msg)
	}
	err := slack.PostWebhookCustomHTTPContext(ctx, webhookURL, c.getHttpClient(), msg)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
