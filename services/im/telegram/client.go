package telegram

import (
	"net/http"
	"time"

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
