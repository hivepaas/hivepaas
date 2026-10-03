package gitea

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A writer is one Gitea answers write, admin or owner for; read, or a user it
// does not know, is not.
func TestCanWrite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/version" {
			_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
			return
		}
		user := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/repos/acme/web/collaborators/"),
			"/permission")
		switch user {
		case "dev":
			_, _ = w.Write([]byte(`{"permission":"write"}`))
		case "lead":
			_, _ = w.Write([]byte(`{"permission":"admin"}`))
		case "reader":
			_, _ = w.Write([]byte(`{"permission":"read"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"user not found"}`))
		}
	}))
	defer srv.Close()

	client, err := NewFromToken("token", srv.URL)
	if !assert.NoError(t, err) {
		return
	}
	for user, want := range map[string]bool{"dev": true, "lead": true, "reader": false, "ghost": false} {
		got, err := client.CanWrite("acme", "web", user)
		assert.NoError(t, err, user)
		assert.Equal(t, want, got, user)
	}
}
