package email

import (
	"context"
	"errors"
	"net/textproto"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/email/http"
	"github.com/hivepaas/hivepaas/services/email/smtp"
)

const (
	defaultEmailRetryMax   = 2
	defaultEmailRetryDelay = 2 * time.Second
)

func SendMail(
	ctx context.Context,
	email *entity.Email,
	recipients []string,
	subject string,
	content string,
) (err error) {
	switch {
	case email.SMTP != nil:
		err = smtp.SendMail(ctx, email.SMTP, recipients, subject, content)
	case email.HTTP != nil:
		err = http.SendMail(ctx, email.HTTP, recipients, subject, content)
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func SendMailRetry(
	ctx context.Context,
	email *entity.Email,
	recipients []string,
	subject string,
	content string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if retryMax <= 0 {
		retryMax = defaultEmailRetryMax
	}
	if retryDelay <= 0 {
		retryDelay = defaultEmailRetryDelay
	}

	for i := range retryMax + 1 {
		if i > 0 {
			timer := time.NewTimer(retryDelay * time.Duration(i))
			select {
			case <-ctx.Done():
				timer.Stop()
				return apperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}

		err = SendMail(ctx, email, recipients, subject, content)
		if err == nil {
			return nil
		}

		if !isRetryableEmailError(email, err) {
			return apperrors.Wrap(err)
		}
	}

	return apperrors.Wrap(err)
}

func isRetryableEmailError(email *entity.Email, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case email != nil && email.SMTP != nil:
		return isRetryableSMTPError(err)
	case email != nil && email.HTTP != nil:
		return isRetryableHTTPError(err)
	default:
		return true
	}
}

//nolint:mnd
func isRetryableSMTPError(err error) bool {
	var protoErr *textproto.Error
	if errors.As(err, &protoErr) {
		if protoErr.Code >= 400 && protoErr.Code < 500 {
			return true
		}
		if protoErr.Code >= 500 {
			return false
		}
	}
	return true
}

func isRetryableHTTPError(err error) bool {
	var httpErr *http.StatusCodeError
	if errors.As(err, &httpErr) {
		return httpErr.Retryable()
	}
	return true
}
