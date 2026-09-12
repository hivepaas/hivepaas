package victorialogs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

func TestPingSucceedsOnHealthyBackend(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, BearerToken: "tok"}})

	assert.NoError(t, c.Ping(context.Background()))
	assert.Equal(t, HealthPath, gotPath)
	assert.Equal(t, "Bearer tok", gotAuth)
}

func TestPingFailsOnUnhealthyBackend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})

	assert.Error(t, c.Ping(context.Background()))
}

func TestIngestURLAppendsTheNativePath(t *testing.T) {
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: "http://vlogs:9428/"}})

	assert.Equal(t, "http://vlogs:9428"+IngestPath, c.IngestURL())
}
