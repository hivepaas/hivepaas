package registryserviceimpl

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushCheckReadsTheAnswer(t *testing.T) {
	tests := []struct {
		name   string
		status int
		ok     bool
		detail string
	}{
		{name: "accepted", status: http.StatusAccepted, ok: true, detail: "went through"},
		{name: "too large", status: http.StatusRequestEntityTooLarge, ok: false, detail: "body limit"},
		{name: "unauthorized", status: http.StatusUnauthorized, ok: false, detail: "credential"},
		{name: "anything else", status: http.StatusBadGateway, ok: false, detail: "502"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readPushCheckAnswer(tt.status)

			assert.Equal(t, tt.ok, result.OK)
			assert.Equal(t, tt.status, result.StatusCode)
			assert.Contains(t, result.Detail, tt.detail)
		})
	}
}

// A registry may hand back a relative Location, and most of them do.
func TestAbsoluteLocation(t *testing.T) {
	assert.Equal(t, "https://registry.example.com/v2/x/blobs/uploads/abc",
		absoluteLocation("registry.example.com", "/v2/x/blobs/uploads/abc"))
	assert.Equal(t, "https://other.example.com/v2/x/blobs/uploads/abc",
		absoluteLocation("registry.example.com", "https://other.example.com/v2/x/blobs/uploads/abc"))
	assert.Empty(t, absoluteLocation("registry.example.com", ""))
}
