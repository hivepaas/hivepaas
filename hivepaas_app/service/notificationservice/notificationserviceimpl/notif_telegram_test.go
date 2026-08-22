package notificationserviceimpl

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/im/telegram"
)

func TestIsRetryableTelegramError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableTelegramError(nil))
	})

	t.Run("general network error", func(t *testing.T) {
		assert.True(t, isRetryableTelegramError(errors.New("connection timeout")))
	})

	t.Run("429 rate limited", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusTooManyRequests,
			Status: "429 Too Many Requests",
		}
		assert.True(t, isRetryableTelegramError(err))
	})

	t.Run("500 internal server error", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusInternalServerError,
			Status: "500 Internal Server Error",
		}
		assert.True(t, isRetryableTelegramError(err))
	})

	t.Run("502 bad gateway", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusBadGateway,
			Status: "502 Bad Gateway",
		}
		assert.True(t, isRetryableTelegramError(err))
	})

	t.Run("400 bad request (non-retryable)", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusBadRequest,
			Status: "400 Bad Request",
		}
		assert.False(t, isRetryableTelegramError(err))
	})

	t.Run("401 unauthorized (non-retryable)", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusUnauthorized,
			Status: "401 Unauthorized",
		}
		assert.False(t, isRetryableTelegramError(err))
	})

	t.Run("403 forbidden (non-retryable)", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusForbidden,
			Status: "403 Forbidden",
		}
		assert.False(t, isRetryableTelegramError(err))
	})

	t.Run("404 not found (non-retryable)", func(t *testing.T) {
		err := &telegram.StatusCodeError{
			Code:   http.StatusNotFound,
			Status: "404 Not Found",
		}
		assert.False(t, isRetryableTelegramError(err))
	})
}
