package notificationserviceimpl

import (
	"errors"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

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
