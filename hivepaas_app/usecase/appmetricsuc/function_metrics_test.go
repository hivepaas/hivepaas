package appmetricsuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/logging"
)

func TestAMetricsWindowEndsWithTheStepNowIsIn(t *testing.T) {
	now := time.Date(2026, 10, 2, 10, 7, 30, 0, time.UTC)
	at := func(month time.Month, day, hour, minute int) time.Time {
		return time.Date(2026, month, day, hour, minute, 0, 0, time.UTC)
	}
	for name, want := range map[string]struct {
		step  time.Duration
		start time.Time
		end   time.Time
	}{
		"1h":  {time.Minute, at(10, 2, 9, 8), at(10, 2, 10, 8)},
		"6h":  {5 * time.Minute, at(10, 2, 4, 10), at(10, 2, 10, 10)},
		"24h": {15 * time.Minute, at(10, 1, 10, 15), at(10, 2, 10, 15)},
		"7d":  {time.Hour, at(9, 25, 11, 0), at(10, 2, 11, 0)},
	} {
		w := metricsWindow(name, now, 0)
		assert.Equal(t, want.step, w.step, name)
		assert.Equal(t, want.start, w.start, name)
		assert.Equal(t, want.end, w.end, name)
		assert.False(t, w.clamped, name)
	}
}

// Logs kept for less than the range answer for what is kept, and say so.
func TestAMetricsWindowStopsWhereTheLogsDo(t *testing.T) {
	now := time.Date(2026, 10, 2, 10, 7, 30, 0, time.UTC)

	w := metricsWindow("7d", now, timeutil.Duration(72*time.Hour))

	assert.Equal(t, time.Date(2026, 9, 29, 11, 0, 0, 0, time.UTC), w.start)
	assert.True(t, w.clamped)
	assert.False(t, metricsWindow("24h", now, timeutil.Duration(72*time.Hour)).clamped)
}

func TestMetricsDataAreTheCountsAsTheAPIWritesThem(t *testing.T) {
	at := time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC)
	p := func(v float64) *float64 { return &v }
	w := metricsWindow("1h", at, 0)

	data := toFunctionMetricsData("1h", w, &logging.InvocationStatsResp{
		Buckets: []*logging.InvocationBucket{
			{Time: at, InvocationCounts: logging.InvocationCounts{Calls: 3, Failed: 1, Errors5xx: 1,
				P50: p(10), P95: p(20), P99: p(30)}},
			{Time: at.Add(time.Minute)},
		},
		Totals: logging.InvocationCounts{Calls: 3, Failed: 1, Errors4xx: 1, Errors5xx: 1,
			P50: p(10), P95: p(20), P99: p(30)},
		ByOutcome: map[string]int64{"ok": 2, "error": 1},
		ByPath: []*logging.InvocationPath{
			{Method: "GET", Path: "/", InvocationCounts: logging.InvocationCounts{Calls: 2, Failed: 1, Errors5xx: 1,
				P95: p(20)}},
			{Method: "GET", Path: "/.env", InvocationCounts: logging.InvocationCounts{Calls: 1, Errors4xx: 1, P95: p(1)}},
		},
	})

	assert.True(t, data.Available)
	assert.Equal(t, "1h", data.Range)
	assert.Equal(t, 60, data.StepSeconds)
	assert.Equal(t, int64(3), data.Totals.Calls)
	assert.Equal(t, int64(1), data.Totals.Errors4xx)
	assert.Equal(t, map[string]int64{"ok": 2, "error": 1}, data.ByOutcome)
	if assert.Len(t, data.ByPath, 2) {
		assert.Equal(t, "GET", data.ByPath[0].Method)
		assert.Equal(t, "/", data.ByPath[0].Path)
		assert.Equal(t, int64(2), data.ByPath[0].Calls)
		assert.Equal(t, "/.env", data.ByPath[1].Path)
		assert.Equal(t, int64(1), data.ByPath[1].Errors4xx)
	}
	if assert.Len(t, data.Series, 2) {
		assert.Equal(t, at, data.Series[0].Time)
		assert.Equal(t, 20.0, *data.Series[0].P95)
		assert.Equal(t, int64(0), data.Series[1].Calls)
		assert.Nil(t, data.Series[1].P95, "no call, no duration")
	}
}
