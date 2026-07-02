package discord

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

type WebhookMessageOption func(webhook *discordgo.WebhookParams)

func (c *Client) WebhookExecute(_ context.Context, webhookURL string, wait bool, content string,
	options ...WebhookMessageOption) (*discordgo.Message, error) {
	webhookID, token, err := parseWebhookURL(webhookURL)
	if err != nil {
		return nil, apperrors.New(err)
	}

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, apperrors.New(err)
	}

	msg := &discordgo.WebhookParams{}
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		if err := json.Unmarshal(reflectutil.UnsafeStrToBytes(content), msg); err != nil {
			msg.Content = content
		}
	} else {
		msg.Content = content
	}

	for _, opt := range options {
		opt(msg)
	}
	resp, err := discord.WebhookExecute(webhookID, token, wait, msg, func(cfg *discordgo.RequestConfig) {
		cfg.Client = c.getHttpClient()
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}
