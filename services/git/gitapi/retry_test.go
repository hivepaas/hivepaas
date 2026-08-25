package gitapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"syscall"
	"testing"
	"time"

	gogithub "github.com/google/go-github/v85/github"
	"github.com/stretchr/testify/assert"
	gogitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestIsRetryableGitError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		retryable, _ := isRetryableGitError(nil)
		assert.False(t, retryable)
	})

	t.Run("context canceled is not retryable", func(t *testing.T) {
		retryable, _ := isRetryableGitError(context.Canceled)
		assert.False(t, retryable)
	})

	t.Run("context deadline exceeded is retryable", func(t *testing.T) {
		retryable, _ := isRetryableGitError(context.DeadlineExceeded)
		assert.True(t, retryable)
	})

	t.Run("network transport errors are retryable", func(t *testing.T) {
		retryable, _ := isRetryableGitError(io.EOF)
		assert.True(t, retryable)

		retryable, _ = isRetryableGitError(syscall.ECONNRESET)
		assert.True(t, retryable)

		retryable, _ = isRetryableGitError(syscall.ECONNREFUSED)
		assert.True(t, retryable)
	})

	t.Run("GitHub 429 Too Many Requests is retryable", func(t *testing.T) {
		err := &gogithub.ErrorResponse{
			Response: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"5"}},
			},
		}
		retryable, retryAfter := isRetryableGitError(err)
		assert.True(t, retryable)
		assert.Equal(t, 5*time.Second, retryAfter)
	})

	t.Run("GitHub AbuseRateLimitError is retryable", func(t *testing.T) {
		retryAfterDur := 10 * time.Second
		err := &gogithub.AbuseRateLimitError{
			RetryAfter: &retryAfterDur,
		}
		retryable, retryAfter := isRetryableGitError(err)
		assert.True(t, retryable)
		assert.Equal(t, 10*time.Second, retryAfter)
	})

	t.Run("GitHub 403 Secondary Rate Limit is retryable", func(t *testing.T) {
		err := &gogithub.ErrorResponse{
			Response: &http.Response{
				StatusCode: http.StatusForbidden,
				Header:     http.Header{"Retry-After": []string{"15"}},
			},
			Message: "You have exceeded a secondary rate limit.",
		}
		retryable, retryAfter := isRetryableGitError(err)
		assert.True(t, retryable)
		assert.Equal(t, 15*time.Second, retryAfter)
	})

	t.Run("GitHub 403 standard permission error is NOT retryable", func(t *testing.T) {
		err := &gogithub.ErrorResponse{
			Response: &http.Response{
				StatusCode: http.StatusForbidden,
				Header:     http.Header{},
			},
			Message: "Resource not accessible by integration",
		}
		retryable, _ := isRetryableGitError(err)
		assert.False(t, retryable)
	})

	t.Run("GitLab 429 is retryable", func(t *testing.T) {
		err := &gogitlab.ErrorResponse{
			Response: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"3"}},
			},
		}
		retryable, retryAfter := isRetryableGitError(err)
		assert.True(t, retryable)
		assert.Equal(t, 3*time.Second, retryAfter)
	})

	t.Run("5xx server errors are retryable", func(t *testing.T) {
		for _, code := range []int{500, 502, 503, 504} {
			err := &gogithub.ErrorResponse{
				Response: &http.Response{StatusCode: code},
			}
			retryable, _ := isRetryableGitError(err)
			assert.True(t, retryable, "expected status %d to be retryable", code)
		}
	})

	t.Run("Client errors (400, 401, 404, 422) are NOT retryable", func(t *testing.T) {
		for _, code := range []int{400, 401, 404, 422, 410} {
			err := &gogithub.ErrorResponse{
				Response: &http.Response{StatusCode: code},
			}
			retryable, _ := isRetryableGitError(err)
			assert.False(t, retryable, "expected status %d NOT to be retryable", code)
		}
	})
}

func TestParseRetryAfterHeader(t *testing.T) {
	assert.Equal(t, 0*time.Second, parseRetryAfterHeader(""))
	assert.Equal(t, 60*time.Second, parseRetryAfterHeader("60"))
	assert.Equal(t, 0*time.Second, parseRetryAfterHeader("invalid-number"))
}

func TestRetryExecute(t *testing.T) {
	ctx := context.Background()

	t.Run("success on first try", func(t *testing.T) {
		attempts := 0
		err := retryExecute(ctx, func() error {
			attempts++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("success on second try after retryable error", func(t *testing.T) {
		attempts := 0
		err := retryExecute(ctx, func() error {
			attempts++
			if attempts == 1 {
				return &gogithub.ErrorResponse{
					Response: &http.Response{StatusCode: http.StatusBadGateway},
				}
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("fails immediately on non-retryable error", func(t *testing.T) {
		attempts := 0
		nonRetryableErr := &gogithub.ErrorResponse{
			Response: &http.Response{StatusCode: http.StatusNotFound},
		}
		err := retryExecute(ctx, func() error {
			attempts++
			return nonRetryableErr
		})
		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("exhausts max attempts on persistent retryable error", func(t *testing.T) {
		attempts := 0
		serverErr := &gogithub.ErrorResponse{
			Response: &http.Response{StatusCode: http.StatusServiceUnavailable},
		}
		err := retryExecute(ctx, func() error {
			attempts++
			return serverErr
		})
		assert.Error(t, err)
		assert.Equal(t, defaultMaxRetryAttempts, attempts)
	})

	t.Run("aborts when context is canceled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		attempts := 0
		err := retryExecute(cancelCtx, func() error {
			attempts++
			return errors.New("something")
		})
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 0, attempts)
	})
}
