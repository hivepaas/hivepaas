package lark

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestGenSign(t *testing.T) {
	sign, err := GenSign("test_secret", 1599360473)
	assert.NoError(t, err)
	assert.NotEmpty(t, sign)

	// Same inputs must yield same signature
	sign2, err := GenSign("test_secret", 1599360473)
	assert.NoError(t, err)
	assert.Equal(t, sign, sign2)
}

func TestStatusCodeError(t *testing.T) {
	t.Run("retryable errors", func(t *testing.T) {
		err429 := &StatusCodeError{Code: http.StatusTooManyRequests, Status: "429 Too Many Requests"}
		assert.True(t, err429.Retryable())

		err500 := &StatusCodeError{Code: http.StatusInternalServerError, Status: "500 Internal Server Error"}
		assert.True(t, err500.Retryable())

		err502 := &StatusCodeError{Code: http.StatusBadGateway, Status: "502 Bad Gateway"}
		assert.True(t, err502.Retryable())

		errLarkRateLimit := &StatusCodeError{Code: 99991400, Status: "request triggered frequency control"}
		assert.True(t, errLarkRateLimit.Retryable())
	})

	t.Run("non-retryable errors", func(t *testing.T) {
		err400 := &StatusCodeError{Code: http.StatusBadRequest, Status: "400 Bad Request"}
		assert.False(t, err400.Retryable())

		err403 := &StatusCodeError{Code: http.StatusForbidden, Status: "403 Forbidden"}
		assert.False(t, err403.Retryable())
	})
}

func TestPostWebhook(t *testing.T) {
	ctx := context.Background()

	t.Run("plain text message without secret", func(t *testing.T) {
		client := &Client{
			httpClient: &http.Client{
				Transport: roundTripFunc(func(r *http.Request) *http.Response {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var payload map[string]any
					err := json.NewDecoder(r.Body).Decode(&payload)
					assert.NoError(t, err)

					assert.Equal(t, "text", payload["msg_type"])
					content, ok := payload["content"].(map[string]any)
					assert.True(t, ok)
					assert.Equal(t, "hello lark", content["text"])
					assert.Nil(t, payload["sign"])
					assert.Nil(t, payload["timestamp"])

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"code": 0, "msg": "success"}`))),
						Header:     make(http.Header),
					}
				}),
			},
		}

		err := client.PostWebhook(ctx, "https://open.larksuite.com/webhook", "", "hello lark")
		assert.NoError(t, err)
	})

	t.Run("json interactive card with secret", func(t *testing.T) {
		client := &Client{
			httpClient: &http.Client{
				Transport: roundTripFunc(func(r *http.Request) *http.Response {
					var payload map[string]any
					err := json.NewDecoder(r.Body).Decode(&payload)
					assert.NoError(t, err)

					assert.Equal(t, "interactive", payload["msg_type"])
					assert.NotEmpty(t, payload["sign"])
					assert.NotEmpty(t, payload["timestamp"])

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"code": 0, "msg": "success"}`))),
						Header:     make(http.Header),
					}
				}),
			},
		}

		err := client.PostWebhook(ctx, "https://open.larksuite.com/webhook", "my_secret",
			`{"msg_type": "interactive", "card": {}}`)
		assert.NoError(t, err)
	})

	t.Run("lark business error response", func(t *testing.T) {
		client := &Client{
			httpClient: &http.Client{
				Transport: roundTripFunc(func(r *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"code": 19001, "msg": "sign match fail"}`))),
						Header:     make(http.Header),
					}
				}),
			},
		}

		err := client.PostWebhook(ctx, "https://open.larksuite.com/webhook", "wrong_secret", "hello")
		assert.Error(t, err)
		var statusErr *StatusCodeError
		assert.ErrorAs(t, err, &statusErr)
		assert.Equal(t, 19001, statusErr.Code)
		assert.Equal(t, "sign match fail", statusErr.Status)
	})
}
