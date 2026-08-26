package imapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	goslack "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/im/lark"
	"github.com/hivepaas/hivepaas/services/im/telegram"
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

func TestIsRetryableDiscordError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableDiscordError(nil))
	})

	t.Run("non-REST general network error", func(t *testing.T) {
		assert.True(t, isRetryableDiscordError(errors.New("connection reset by peer")))
	})

	t.Run("429 rate limited", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusTooManyRequests,
			},
		}
		assert.True(t, isRetryableDiscordError(err))
	})

	t.Run("500 internal server error", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
		}
		assert.True(t, isRetryableDiscordError(err))
	})

	t.Run("502 bad gateway", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusBadGateway,
			},
		}
		assert.True(t, isRetryableDiscordError(err))
	})

	t.Run("400 bad request (non-retryable)", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
		}
		assert.False(t, isRetryableDiscordError(err))
	})

	t.Run("401 unauthorized (non-retryable)", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusUnauthorized,
			},
		}
		assert.False(t, isRetryableDiscordError(err))
	})

	t.Run("404 not found (non-retryable)", func(t *testing.T) {
		err := &discordgo.RESTError{
			Response: &http.Response{
				StatusCode: http.StatusNotFound,
			},
		}
		assert.False(t, isRetryableDiscordError(err))
	})
}

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

func TestIsRetryableLarkError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableLarkError(nil))
	})

	t.Run("general network error", func(t *testing.T) {
		assert.True(t, isRetryableLarkError(errors.New("connection timeout")))
	})

	t.Run("429 rate limited", func(t *testing.T) {
		err := &lark.StatusCodeError{
			Code:   http.StatusTooManyRequests,
			Status: "429 Too Many Requests",
		}
		assert.True(t, isRetryableLarkError(err))
	})

	t.Run("500 internal server error", func(t *testing.T) {
		err := &lark.StatusCodeError{
			Code:   http.StatusInternalServerError,
			Status: "500 Internal Server Error",
		}
		assert.True(t, isRetryableLarkError(err))
	})

	t.Run("502 bad gateway", func(t *testing.T) {
		err := &lark.StatusCodeError{
			Code:   http.StatusBadGateway,
			Status: "502 Bad Gateway",
		}
		assert.True(t, isRetryableLarkError(err))
	})

	t.Run("400 bad request (non-retryable)", func(t *testing.T) {
		err := &lark.StatusCodeError{
			Code:   http.StatusBadRequest,
			Status: "400 Bad Request",
		}
		assert.False(t, isRetryableLarkError(err))
	})

	t.Run("403 forbidden (non-retryable)", func(t *testing.T) {
		err := &lark.StatusCodeError{
			Code:   http.StatusForbidden,
			Status: "403 Forbidden",
		}
		assert.False(t, isRetryableLarkError(err))
	})
}

func TestIsRetryableIMError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableIMError(base.IMServiceKindSlack, nil))
		assert.False(t, isRetryableIMError(base.IMServiceKindLark, nil))
	})

	t.Run("context canceled", func(t *testing.T) {
		assert.False(t, isRetryableIMError(base.IMServiceKindSlack, context.Canceled))
		assert.False(t, isRetryableIMError(base.IMServiceKindDiscord, context.Canceled))
		assert.False(t, isRetryableIMError(base.IMServiceKindTelegram, context.Canceled))
		assert.False(t, isRetryableIMError(base.IMServiceKindLark, context.Canceled))
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		assert.True(t, isRetryableIMError(base.IMServiceKindSlack, context.DeadlineExceeded))
		assert.True(t, isRetryableIMError(base.IMServiceKindDiscord, context.DeadlineExceeded))
		assert.True(t, isRetryableIMError(base.IMServiceKindTelegram, context.DeadlineExceeded))
		assert.True(t, isRetryableIMError(base.IMServiceKindLark, context.DeadlineExceeded))
	})

	t.Run("unknown provider", func(t *testing.T) {
		assert.False(t, isRetryableIMError("unknown", errors.New("some error")))
	})
}

func TestRetryExecute(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		attempts := 0
		err := retryExecute(context.Background(), func() error {
			attempts++
			return nil
		}, base.IMServiceKindSlack, 2, 10*time.Millisecond)

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("success after 1 retry", func(t *testing.T) {
		attempts := 0
		err := retryExecute(context.Background(), func() error {
			attempts++
			if attempts == 1 {
				return errors.New("temporary error")
			}
			return nil
		}, base.IMServiceKindSlack, 2, 10*time.Millisecond)

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("exhaust all retries with retryable error", func(t *testing.T) {
		attempts := 0
		expectedErr := errors.New("network error")
		err := retryExecute(context.Background(), func() error {
			attempts++
			return expectedErr
		}, base.IMServiceKindSlack, 2, 5*time.Millisecond)

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 3, attempts) // 1 initial + 2 retries = 3 attempts total
	})

	t.Run("stops immediately on non-retryable error", func(t *testing.T) {
		attempts := 0
		nonRetryableErr := &goslack.StatusCodeError{
			Code:   http.StatusBadRequest,
			Status: "400 Bad Request",
		}
		err := retryExecute(context.Background(), func() error {
			attempts++
			return nonRetryableErr
		}, base.IMServiceKindSlack, 3, 10*time.Millisecond)

		assert.ErrorIs(t, err, nonRetryableErr)
		assert.Equal(t, 1, attempts)
	})

	t.Run("aborts on context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := retryExecute(ctx, func() error {
			return nil
		}, base.IMServiceKindSlack, 2, 10*time.Millisecond)

		assert.Error(t, err)
	})
}
