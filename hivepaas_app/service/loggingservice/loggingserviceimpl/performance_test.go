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

// obiBackend answers OBIStats with nothing, and keeps the request.
type obiBackend struct {
	logging.Backend
	got *logging.OBIStatsReq
}

func (b *obiBackend) OBIStats(_ context.Context, req *logging.OBIStatsReq) (*logging.OBIStatsResp, error) {
	b.got = req
	return &logging.OBIStatsResp{}, nil
}

// An app's routes and calls are the agent's rows - by the identity the daemon
// wrote - with the app's id, every bucket summed.
func TestOBIMetricsAreTheAgentsRowsForTheApp(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	b := &obiBackend{}
	s.newBackend = func(logging.BackendType, *logging.BackendConfig) (logging.Backend, error) { return b, nil }
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	q := &loggingservice.FunctionMetricsQuery{Start: at, End: at.Add(time.Hour), Step: time.Minute}

	_, err := s.RouteMetrics(context.Background(), nil, &entity.App{ID: "01K6A"}, q)
	assert.NoError(t, err)
	if assert.NotNil(t, b.got) {
		assert.Equal(t, []logging.FieldMatch{{Field: "attrs.hivepaas.component", Value: "agent"}}, b.got.Match)
		assert.Equal(t, "01K6A", b.got.AppID)
		assert.Equal(t, "routes", b.got.Row)
		assert.Equal(t, []string{"kind", "method", "route"}, b.got.GroupBy)
		assert.Len(t, b.got.BucketFields, 12)
		assert.Equal(t, time.Minute, b.got.Step)
	}
	_, err = s.DependencyMetrics(context.Background(), nil, &entity.App{ID: "01K6A"}, q)
	assert.NoError(t, err)
	assert.Equal(t, "calls", b.got.Row)
	assert.Equal(t, []string{"kind"}, b.got.SeriesBy)

	_, err = s.RouteMetrics(context.Background(), nil, &entity.App{}, q)
	assert.ErrorIs(t, err, logging.ErrQueryScopeRequired)
}

// Each node's latest status counts, the rows the latest first; what is no
// status is skipped.
func TestLatestStatuses(t *testing.T) {
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	got := latestStatuses([]*logging.OBIStatusRow{
		{Time: at.Add(2 * time.Minute), Msg: `{"hp":"obi","node":"n1","running":true,"preflight":{"ok":true}}`},
		{Time: at.Add(time.Minute), Msg: `{"hp":"obi","node":"n2","preflight":{"reasons":["no-btf"]}}`},
		{Time: at, Msg: `{"hp":"obi","node":"n1","running":false}`},
		{Time: at, Msg: `{"hp":"routes","node":"n3"}`},
		{Time: at, Msg: `{"hp":"obi"}`},
		{Time: at, Msg: `not json`},
	})
	assert.Len(t, got, 2)
	assert.True(t, got["n1"].Running)
	assert.Equal(t, at.Add(2*time.Minute), got["n1"].Time)
	assert.Equal(t, []string{"no-btf"}, got["n2"].Preflight.Reasons)
}
