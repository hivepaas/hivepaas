package loggingserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// statsBackend answers InvocationStats with resp, and keeps the request.
type statsBackend struct {
	logging.Backend
	got  *logging.InvocationStatsReq
	resp *logging.InvocationStatsResp
}

func (b *statsBackend) InvocationStats(
	_ context.Context, req *logging.InvocationStatsReq,
) (*logging.InvocationStatsResp, error) {
	b.got = req
	return b.resp, nil
}

func withStatsBackend(s *service, resp *logging.InvocationStatsResp) *statsBackend {
	b := &statsBackend{resp: resp}
	s.newBackend = func(logging.BackendType, *logging.BackendConfig) (logging.Backend, error) { return b, nil }
	return b
}

// A function's metrics are scoped to it, and every step of the range is a
// point: one without a call is one with no calls.
func TestFunctionMetricsAreTheFunctionsAndEveryStepIsAPoint(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	at := func(m int) time.Time { return time.Date(2026, 10, 2, 10, m, 0, 0, time.UTC) }
	p := func(v float64) *float64 { return &v }
	b := withStatsBackend(s, &logging.InvocationStatsResp{
		Buckets: []*logging.InvocationBucket{
			{Time: at(1), InvocationCounts: logging.InvocationCounts{Calls: 2, P95: p(12)}},
			{Time: at(3), InvocationCounts: logging.InvocationCounts{Calls: 1, Failed: 1, P95: p(40)}},
		},
		Totals:    logging.InvocationCounts{Calls: 3, Failed: 1, P95: p(40)},
		ByOutcome: map[string]int64{"ok": 2, "error": 1},
	})

	got, err := s.FunctionMetrics(context.Background(), nil, &entity.App{ID: "FN1"},
		&loggingservice.FunctionMetricsQuery{Start: at(0), End: at(5), Step: time.Minute})

	assert.NoError(t, err)
	if assert.NotNil(t, b.got) {
		assert.Equal(t, []logging.FieldMatch{{Field: "attrs." + appservice.LabelLogAppID, Value: "FN1"}}, b.got.Match)
		assert.Equal(t, time.Minute, b.got.Step)
	}
	if assert.Len(t, got.Buckets, 5) {
		for i, bucket := range got.Buckets {
			assert.Equal(t, at(i), bucket.Time)
		}
		assert.Equal(t, int64(0), got.Buckets[0].Calls)
		assert.Nil(t, got.Buckets[0].P95)
		assert.Equal(t, int64(2), got.Buckets[1].Calls)
		assert.Equal(t, int64(1), got.Buckets[3].Failed)
	}
	assert.Equal(t, int64(3), got.Totals.Calls)
	assert.Equal(t, map[string]int64{"ok": 2, "error": 1}, got.ByOutcome)
}

func TestFunctionMetricsRefuseAnAppWithoutAnID(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	b := withStatsBackend(s, &logging.InvocationStatsResp{})

	_, err := s.FunctionMetrics(context.Background(), nil, &entity.App{},
		&loggingservice.FunctionMetricsQuery{Start: time.Now().Add(-time.Hour), End: time.Now(), Step: time.Minute})

	assert.Error(t, err)
	assert.Nil(t, b.got, "an empty id would count every line with no app at all")
}
