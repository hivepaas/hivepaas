package loggingserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// httpBackend answers HTTPStats with resp, and keeps the request.
type httpBackend struct {
	logging.Backend
	got  *logging.HTTPStatsReq
	resp *logging.HTTPStatsResp
}

func (b *httpBackend) HTTPStats(_ context.Context, req *logging.HTTPStatsReq) (*logging.HTTPStatsResp, error) {
	b.got = req
	return b.resp, nil
}

// An app's HTTP numbers are the proxy's lines - by the identity the daemon
// wrote - for the app's services, by its id; every step of the range a point.
func TestHTTPMetricsAreTheProxysLinesForTheApp(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	at := func(m int) time.Time { return time.Date(2026, 10, 2, 10, m, 0, 0, time.UTC) }
	b := &httpBackend{resp: &logging.HTTPStatsResp{
		Buckets: []*logging.HTTPBucket{{Time: at(2), HTTPCounts: logging.HTTPCounts{Requests: 4, Errors5xx: 1}}},
		Totals:  logging.HTTPCounts{Requests: 4, Errors5xx: 1},
	}}
	s.newBackend = func(logging.BackendType, *logging.BackendConfig) (logging.Backend, error) { return b, nil }

	got, err := s.HTTPMetrics(context.Background(), nil, &entity.App{ID: "01K6A"},
		&loggingservice.FunctionMetricsQuery{Start: at(0), End: at(4), Step: time.Minute})

	assert.NoError(t, err)
	if assert.NotNil(t, b.got) {
		assert.Equal(t, []logging.FieldMatch{{Field: "attrs.hivepaas.component", Value: "traefik"}}, b.got.Match)
		assert.Equal(t, "^svc-01k6a-[0-9]+@swarm$", b.got.ServicePattern)
		assert.Equal(t, 20, b.got.TopPaths)
	}
	if assert.Len(t, got.Buckets, 4) {
		assert.Equal(t, int64(0), got.Buckets[0].Requests)
		assert.Equal(t, int64(4), got.Buckets[2].Requests)
	}

	_, err = s.HTTPMetrics(context.Background(), nil, &entity.App{},
		&loggingservice.FunctionMetricsQuery{Start: at(0), End: at(4), Step: time.Minute})
	assert.ErrorIs(t, err, logging.ErrQueryScopeRequired)
}

// The proxy's lines are collected with the apps' or with HivePaaS's; the app's
// own log driver does not matter.
func TestProxyHistory(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	h, err := s.ProxyHistory(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, h.Available)

	cfg := enabledConfig()
	cfg.Sources = entity.LoggingSources{}
	s = newTestService(&fakeDocker{}, storedSetting(t, cfg))
	h, err = s.ProxyHistory(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, loggingservice.HistoryReasonAppsNotCollected, h.Reason)

	s = newTestService(&fakeDocker{}, nil)
	h, err = s.ProxyHistory(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, loggingservice.HistoryReasonDisabled, h.Reason)
}
