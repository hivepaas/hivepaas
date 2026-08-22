package notificationserviceimpl

import (
	"errors"
	"net/http"
	"testing"
	"time"

	goslack "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestIsRetryableSlackError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableSlackError(nil))
	})

	t.Run("general network error", func(t *testing.T) {
		assert.True(t, isRetryableSlackError(errors.New("connection timeout")))
	})

	t.Run("rate limited error", func(t *testing.T) {
		err := &goslack.RateLimitedError{
			RetryAfter: 2 * time.Second,
		}
		assert.True(t, isRetryableSlackError(err))
	})

	t.Run("500 internal server error", func(t *testing.T) {
		err := &goslack.StatusCodeError{
			Code:   http.StatusInternalServerError,
			Status: "500 Internal Server Error",
		}
		assert.True(t, isRetryableSlackError(err))
	})

	t.Run("502 bad gateway", func(t *testing.T) {
		err := &goslack.StatusCodeError{
			Code:   http.StatusBadGateway,
			Status: "502 Bad Gateway",
		}
		assert.True(t, isRetryableSlackError(err))
	})

	t.Run("400 bad request (non-retryable)", func(t *testing.T) {
		err := &goslack.StatusCodeError{
			Code:   http.StatusBadRequest,
			Status: "400 Bad Request",
		}
		assert.False(t, isRetryableSlackError(err))
	})

	t.Run("403 forbidden (non-retryable)", func(t *testing.T) {
		err := &goslack.StatusCodeError{
			Code:   http.StatusForbidden,
			Status: "403 Forbidden",
		}
		assert.False(t, isRetryableSlackError(err))
	})

	t.Run("404 not found (non-retryable)", func(t *testing.T) {
		err := &goslack.StatusCodeError{
			Code:   http.StatusNotFound,
			Status: "404 Not Found",
		}
		assert.False(t, isRetryableSlackError(err))
	})
}
