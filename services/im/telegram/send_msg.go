package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type StatusCodeError struct {
	Code   int
	Status string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("telegram server error: %s", e.Status)
}

func (e *StatusCodeError) Retryable() bool {
	return e.Code >= http.StatusInternalServerError || e.Code == http.StatusTooManyRequests
}

type sendMessagePayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (c *Client) SendMessage(ctx context.Context, botToken, chatID, text, parseMode string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := sendMessagePayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return apperrors.Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return apperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.getHttpClient().Do(req)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return &StatusCodeError{Code: resp.StatusCode, Status: resp.Status}
	}

	return nil
}
