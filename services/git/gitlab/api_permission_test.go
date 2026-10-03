package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A writer is a member at Developer or above, directly or by a group; a
// Reporter, or someone who is not a member, is not.
func TestCanWrite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.EscapedPath() {
		case "/api/v4/projects/acme%2Fweb/members/all/1":
			_, _ = w.Write([]byte(`{"id":1,"access_level":30}`))
		case "/api/v4/projects/acme%2Fweb/members/all/2":
			_, _ = w.Write([]byte(`{"id":2,"access_level":20}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"404 Not found"}`))
		}
	}))
	defer srv.Close()

	client, err := NewFromToken("token", srv.URL)
	if !assert.NoError(t, err) {
		return
	}
	for id, want := range map[int64]bool{1: true, 2: false, 3: false} {
		got, err := client.CanWrite(context.Background(), "acme/web", id)
		assert.NoError(t, err, id)
		assert.Equal(t, want, got, id)
	}
}
