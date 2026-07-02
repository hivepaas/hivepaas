package discord

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/httpclient"
)

const (
	defaultHttpClientTimeout = 10 * time.Second
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) getHttpClient() *http.Client {
	if c.httpClient == nil {
		return &http.Client{
			Timeout:   defaultHttpClientTimeout,
			Transport: httpclient.DefaultClient.Transport,
		}
	}
	return c.httpClient
}

func parseWebhookURL(webhook string) (webhookID, token string, err error) {
	dcUrl, err := url.Parse(webhook)
	if err != nil {
		return "", "", apperrors.New(err)
	}
	parts := strings.Split(dcUrl.Path, "/")
	if len(parts) > 2 { //nolint:mnd
		return parts[len(parts)-2], parts[len(parts)-1], nil
	}
	return "", "", apperrors.New(apperrors.ErrArgumentInvalid).
		WithMsgLog("unabled to parse webhook URL")
}
