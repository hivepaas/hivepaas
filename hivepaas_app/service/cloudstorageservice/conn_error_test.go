package cloudstorageservice

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// statusError answers as the SDK's response errors do.
type statusError struct {
	code  int
	cause error
}

func (e *statusError) Error() string       { return fmt.Sprintf("response error StatusCode: %d", e.code) }
func (e *statusError) HTTPStatusCode() int { return e.code }
func (e *statusError) Unwrap() error       { return e.cause }

func TestConnErrorSaysWhatToCheck(t *testing.T) {
	dns := &net.DNSError{Err: "no such host", Name: "bucket.s3.example.invalid", IsNotFound: true}
	for want, err := range map[string]error{
		"the key was refused":               &statusError{code: 403},
		"no such bucket":                    &statusError{code: 404},
		"check the region":                  &statusError{code: 301},
		"the store answered Internal":       &statusError{code: 500},
		"the endpoint could not be reached": &statusError{code: 0, cause: dns},
		"check the endpoint, the region":    errors.New("something else"),
	} {
		wrapped := fmt.Errorf("operation error S3: HeadBucket: %w", err)
		assert.ErrorIs(t, ConnError(wrapped), hperrors.ErrCloudStorageConnFailed, want)
		assert.Contains(t, connReason(wrapped), want)
	}
	assert.NoError(t, ConnError(nil))
}
