package hperrors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A key nobody knows is the caller not being authenticated: 401, as no
// credentials at all are - not a precondition the request failed.
func TestAnInvalidAPIKeyIsUnauthorized(t *testing.T) {
	assert.ErrorIs(t, ErrAPIKeyInvalid, ErrUnauthorized)
	assert.Equal(t, http.StatusUnauthorized, Wrap(ErrAPIKeyInvalid).StatusCode())
}
