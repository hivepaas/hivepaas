package email

import (
	"errors"
	nethttp "net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/email/http"
)

func TestIsRetryableSMTPError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isRetryableEmailError(&entity.Email{SMTP: &entity.EmailSMTP{}}, nil))
	})

	t.Run("general network error", func(t *testing.T) {
		assert.True(t, isRetryableSMTPError(errors.New("dial tcp: i/o timeout")))
	})

	t.Run("SMTP 421 server busy", func(t *testing.T) {
		err := &textproto.Error{Code: 421, Msg: "Service not available"}
		assert.True(t, isRetryableSMTPError(err))
	})

	t.Run("SMTP 450 mailbox busy", func(t *testing.T) {
		err := &textproto.Error{Code: 450, Msg: "Mailbox unavailable, busy"}
		assert.True(t, isRetryableSMTPError(err))
	})

	t.Run("SMTP 535 auth failed (non-retryable)", func(t *testing.T) {
		err := &textproto.Error{Code: 535, Msg: "Authentication credentials invalid"}
		assert.False(t, isRetryableSMTPError(err))
	})

	t.Run("SMTP 550 user not found (non-retryable)", func(t *testing.T) {
		err := &textproto.Error{Code: 550, Msg: "Requested action not taken: mailbox unavailable"}
		assert.False(t, isRetryableSMTPError(err))
	})

	t.Run("SMTP 554 spam rejection (non-retryable)", func(t *testing.T) {
		err := &textproto.Error{Code: 554, Msg: "Transaction failed, spam detected"}
		assert.False(t, isRetryableSMTPError(err))
	})
}

func TestIsRetryableHTTPError(t *testing.T) {
	t.Run("general network error", func(t *testing.T) {
		assert.True(t, isRetryableHTTPError(errors.New("connection reset by peer")))
	})

	t.Run("429 rate limited", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusTooManyRequests,
			Status: "429 Too Many Requests",
		}
		assert.True(t, isRetryableHTTPError(err))
	})

	t.Run("500 internal server error", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusInternalServerError,
			Status: "500 Internal Server Error",
		}
		assert.True(t, isRetryableHTTPError(err))
	})

	t.Run("502 bad gateway", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusBadGateway,
			Status: "502 Bad Gateway",
		}
		assert.True(t, isRetryableHTTPError(err))
	})

	t.Run("400 bad request (non-retryable)", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusBadRequest,
			Status: "400 Bad Request",
		}
		assert.False(t, isRetryableHTTPError(err))
	})

	t.Run("401 unauthorized (non-retryable)", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusUnauthorized,
			Status: "401 Unauthorized",
		}
		assert.False(t, isRetryableHTTPError(err))
	})

	t.Run("404 not found (non-retryable)", func(t *testing.T) {
		err := &http.StatusCodeError{
			Code:   nethttp.StatusNotFound,
			Status: "404 Not Found",
		}
		assert.False(t, isRetryableHTTPError(err))
	})
}

func TestIsRetryableEmailError_Dispatch(t *testing.T) {
	t.Run("dispatch to SMTP", func(t *testing.T) {
		emailSMTP := &entity.Email{SMTP: &entity.EmailSMTP{}}
		err535 := &textproto.Error{Code: 535, Msg: "Auth failed"}
		err421 := &textproto.Error{Code: 421, Msg: "Server busy"}

		assert.False(t, isRetryableEmailError(emailSMTP, err535))
		assert.True(t, isRetryableEmailError(emailSMTP, err421))
	})

	t.Run("dispatch to HTTP", func(t *testing.T) {
		emailHTTP := &entity.Email{HTTP: &entity.EmailHTTP{}}
		err401 := &http.StatusCodeError{Code: 401}
		err500 := &http.StatusCodeError{Code: 500}

		assert.False(t, isRetryableEmailError(emailHTTP, err401))
		assert.True(t, isRetryableEmailError(emailHTTP, err500))
	})
}
