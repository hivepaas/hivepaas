package victorialogs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func TestQuerySendsTheBuiltQueryAndTimeRange(t *testing.T) {
	var got url.Values
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, QueryPath, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		got, _ = url.ParseQuery(string(body))
		auth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w,
			`{"_time":"2026-09-12T08:00:02.5Z","_msg":"second","stream":"stderr","app.level":"error"}`+"\n"+
				`{"_time":"2026-09-12T08:00:01Z","_msg":"first","stream":"stdout"}`+"\n")
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, BearerToken: "tok"}})
	start := time.Date(2026, 9, 12, 7, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)
	resp, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "attrs.hivepaas.app.id", Value: "APP1"}},
		Start: start, End: end, Limit: 2,
	})
	assert.NoError(t, err)

	want, _ := BuildQuery(&loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "attrs.hivepaas.app.id", Value: "APP1"}}, Limit: 2,
	})
	assert.Equal(t, want, got.Get("query"))
	assert.Equal(t, "2026-09-12T07:00:00Z", got.Get("start"))
	assert.Equal(t, "2026-09-12T09:00:00Z", got.Get("end"))
	assert.Equal(t, "Bearer tok", auth)

	if assert.Len(t, resp.Entries, 2) {
		assert.Equal(t, "first", resp.Entries[0].Message, "oldest first")
		assert.Equal(t, "stdout", resp.Entries[0].Stream)
		assert.Equal(t, "", resp.Entries[0].Level)
		assert.Equal(t, "second", resp.Entries[1].Message)
		assert.Equal(t, "error", resp.Entries[1].Level)
		assert.Equal(t, 500*time.Millisecond, resp.Entries[1].Time.Sub(resp.Entries[0].Time)-time.Second)
	}
	assert.True(t, resp.Truncated, "as many lines as the limit means older ones may exist")
}

func TestQueryMapsABadRequestToQueryInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "cannot parse query", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})
	_, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1,
	})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
}

func TestQueryMapsAnUnreachableBackend(t *testing.T) {
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: "http://127.0.0.1:1"}})
	_, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1,
	})
	assert.ErrorIs(t, err, loggingmodel.ErrBackendUnreachable)
}

func TestQueryHonorsTLSSkipVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "")
	}))
	defer srv.Close()
	req := &loggingmodel.QueryReq{Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1}

	_, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}}).Query(context.Background(), req)
	assert.ErrorIs(t, err, loggingmodel.ErrBackendUnreachable, "a self-signed cert is refused by default")

	resp, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, TLSSkipVerify: true}}).
		Query(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, resp.Entries)
	assert.False(t, resp.Truncated)
}

func TestQueryRefusesAnUnscopedRequestWithoutCalling(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true }))
	defer srv.Close()

	_, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}}).
		Query(context.Background(), &loggingmodel.QueryReq{Limit: 1})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	assert.False(t, called)
}
