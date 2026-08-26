package imapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	goslack "github.com/slack-go/slack"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/services/im/lark"
	"github.com/hivepaas/hivepaas/services/im/telegram"
)

const (
	defaultInitialDelay = 2 * time.Second
	defaultDelayIncr    = 1 * time.Second
)

// retryExecute executes an operation with smart retry on transient errors.
func retryExecute(
	ctx context.Context,
	op func() error,
	imKind base.IMServiceKind,
	retryMax int,
	retryDelay time.Duration,
) error {
	var lastErr error
	if retryDelay <= 0 {
		retryDelay = defaultInitialDelay
	}
	delayIncr := defaultDelayIncr

	for attempt := 0; attempt <= retryMax; attempt++ {
		if ctx.Err() != nil {
			return apperrors.Wrap(ctx.Err())
		}

		err := op()
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt == retryMax {
			break
		}

		retryable := isRetryableIMError(imKind, err)
		if !retryable {
			return err
		}

		delay := executil.RetryDelay(attempt, retryDelay, &delayIncr, nil)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return apperrors.Wrap(ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

// isRetryableIMError determines whether a given error from a IM provider API should be retried.
func isRetryableIMError(imKind base.IMServiceKind, err error) bool {
	if err == nil {
		return false
	}

	// 1. Context errors
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	switch imKind {
	case base.IMServiceKindSlack:
		return isRetryableSlackError(err)
	case base.IMServiceKindDiscord:
		return isRetryableDiscordError(err)
	case base.IMServiceKindTelegram:
		return isRetryableTelegramError(err)
	case base.IMServiceKindLark:
		return isRetryableLarkError(err)
	}
	return false
}

func isRetryableDiscordError(err error) bool {
	if err == nil {
		return false
	}
	var restErr *discordgo.RESTError
	if errors.As(err, &restErr) && restErr.Response != nil {
		statusCode := restErr.Response.StatusCode
		if statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError {
			return true
		}
		return false
	}
	return true
}

func isRetryableSlackError(err error) bool {
	if err == nil {
		return false
	}
	var rateLimitedErr *goslack.RateLimitedError
	if errors.As(err, &rateLimitedErr) {
		return true
	}
	var statusCodeErr *goslack.StatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.Retryable()
	}
	return true
}

func isRetryableTelegramError(err error) bool {
	if err == nil {
		return false
	}
	var statusCodeErr *telegram.StatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.Retryable()
	}
	return true
}

func isRetryableLarkError(err error) bool {
	if err == nil {
		return false
	}
	var statusCodeErr *lark.StatusCodeError
	if errors.As(err, &statusCodeErr) {
		return statusCodeErr.Retryable()
	}
	return true
}
