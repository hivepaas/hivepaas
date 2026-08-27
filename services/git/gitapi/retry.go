package gitapi

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"

	gogithub "github.com/google/go-github/v85/github"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
)

const (
	defaultMaxRetryAttempts = 3
	defaultInitialDelay     = 1 * time.Second
	defaultJitter           = 500 * time.Millisecond
	maxRetryAfterCap        = 30 * time.Second
)

// retryExecute executes an operation with smart retry on transient errors.
func retryExecute(ctx context.Context, op func() error) error {
	var lastErr error
	jitter := defaultJitter

	for attempt := 1; attempt <= defaultMaxRetryAttempts; attempt++ {
		if ctx.Err() != nil {
			return hperrors.Wrap(ctx.Err())
		}

		err := op()
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt == defaultMaxRetryAttempts {
			break
		}

		retryable, retryAfter := isRetryableGitError(err)
		if !retryable {
			return err
		}

		var delay time.Duration
		if retryAfter > 0 {
			delay = retryAfter
			if delay > maxRetryAfterCap {
				delay = maxRetryAfterCap
			}
		} else {
			delay = executil.RetryDelay(attempt, defaultInitialDelay, nil, &jitter)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return hperrors.Wrap(ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

// isRetryableGitError determines whether a given error from a Git provider API should be retried.
func isRetryableGitError(err error) (retryable bool, retryAfter time.Duration) {
	if err == nil {
		return false, 0
	}

	// 1. Context errors
	if errors.Is(err, context.Canceled) {
		return false, 0
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true, 0
	}

	// 2. Network & Transport level errors
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true, 0
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ETIMEDOUT) {
		return true, 0
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true, 0
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return true, 0
		}
		if urlErr.Err != nil {
			return isRetryableGitError(urlErr.Err)
		}
	}

	// 3. GitHub API Errors
	var ghAbuseErr *gogithub.AbuseRateLimitError
	if errors.As(err, &ghAbuseErr) {
		return true, ghAbuseErr.GetRetryAfter()
	}

	var ghRateLimitErr *gogithub.RateLimitError
	if errors.As(err, &ghRateLimitErr) {
		if !ghRateLimitErr.Rate.Reset.IsZero() {
			diff := time.Until(ghRateLimitErr.Rate.Reset.Time)
			if diff > 0 {
				return true, diff
			}
		}
		return true, 0
	}

	var ghErrResp *gogithub.ErrorResponse
	if errors.As(err, &ghErrResp) && ghErrResp.Response != nil {
		return handleHTTPStatusCode(ghErrResp.Response.StatusCode, ghErrResp.Response.Header, ghErrResp.Message)
	}

	// 4. GitLab API Errors
	var glErrResp *gogitlab.ErrorResponse
	if errors.As(err, &glErrResp) && glErrResp.Response != nil {
		return handleHTTPStatusCode(glErrResp.Response.StatusCode, glErrResp.Response.Header, glErrResp.Message)
	}

	// 5. Fallback for wrapped errors containing HTTP status
	msg := err.Error()
	if strings.Contains(msg, "429") ||
		strings.Contains(msg, "502 Bad Gateway") ||
		strings.Contains(msg, "503 Service Unavailable") ||
		strings.Contains(msg, "504 Gateway Timeout") ||
		strings.Contains(msg, "secondary rate limit") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "connection refused") {
		return true, 0
	}

	return false, 0
}

func handleHTTPStatusCode(statusCode int, header http.Header, msg string) (bool, time.Duration) {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		var retryAfter time.Duration
		if header != nil {
			retryAfter = parseRetryAfterHeader(header.Get("Retry-After"))
		}
		return true, retryAfter

	case http.StatusForbidden: // 403
		// Check if 403 is due to GitHub secondary rate limit
		if header != nil && header.Get("Retry-After") != "" {
			return true, parseRetryAfterHeader(header.Get("Retry-After"))
		}
		if strings.Contains(strings.ToLower(msg), "rate limit") ||
			strings.Contains(strings.ToLower(msg), "secondary rate limit") {
			return true, 0
		}
		return false, 0

	case http.StatusRequestTimeout, // 408
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		var retryAfter time.Duration
		if header != nil {
			retryAfter = parseRetryAfterHeader(header.Get("Retry-After"))
		}
		return true, retryAfter

	default:
		// 400, 401, 404, 422, 410, etc. are non-retryable
		return false, 0
	}
}

func parseRetryAfterHeader(headerVal string) time.Duration {
	if headerVal == "" {
		return 0
	}

	// 1. Integer seconds (e.g. "120")
	if seconds, err := strconv.Atoi(strings.TrimSpace(headerVal)); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// 2. HTTP-Date (e.g. "Wed, 21 Oct 2025 07:28:00 GMT")
	if t, err := http.ParseTime(headerVal); err == nil {
		diff := time.Until(t)
		if diff > 0 {
			return diff
		}
	}

	return 0
}
