package lark

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

type StatusCodeError struct {
	Code   int
	Status string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("lark server error: code=%d, msg=%s", e.Code, e.Status)
}

func (e *StatusCodeError) Retryable() bool {
	return e.Code == http.StatusTooManyRequests ||
		e.Code >= http.StatusInternalServerError ||
		e.Code == 99991400 // Lark rate limit error code
}

type larkResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GenSign generates a signature for Lark custom bot webhook.
// stringToSign = timestamp + "\n" + secret
func GenSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v\n%v", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write([]byte{})
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

//nolint:goconst
func (c *Client) PostWebhook(ctx context.Context, webhookURL, secret, text string) error {
	trimmed := strings.TrimSpace(text)
	var payload map[string]any

	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		if err := json.Unmarshal(reflectutil.UnsafeStrToBytes(trimmed), &payload); err != nil {
			payload = map[string]any{
				"msg_type": "text",
				"content": map[string]any{
					"text": text,
				},
			}
		}
	} else {
		payload = map[string]any{
			"msg_type": "text",
			"content": map[string]any{
				"text": text,
			},
		}
	}

	if secret != "" {
		timestamp := time.Now().Unix()
		sign, err := GenSign(secret, timestamp)
		if err != nil {
			return apperrors.Wrap(err)
		}
		payload["timestamp"] = strconv.FormatInt(timestamp, 10)
		payload["sign"] = sign
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return apperrors.Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return apperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.getHttpClient().Do(req)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		return &StatusCodeError{
			Code:   resp.StatusCode,
			Status: resp.Status,
		}
	}

	var result larkResp
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return &StatusCodeError{
			Code:   result.Code,
			Status: result.Msg,
		}
	}

	return nil
}
