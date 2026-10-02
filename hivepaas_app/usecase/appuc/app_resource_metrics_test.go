package appuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging"
)

// Totals: CPU's average over the steps with a row and its peak, memory's peak,
// the limits last seen, the OOM kills summed.
func TestResourceTotals(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	data := toAppResourceMetricsData("1h", window{start: at, end: at.Add(3 * time.Minute), step: time.Minute},
		&logging.ResourceStatsResp{Buckets: []*logging.ResourceBucket{
			{Time: at, ResourceUsage: logging.ResourceUsage{CPU: f(0.2), CPULimit: 1, Memory: f(100), MemoryLimit: 512}},
			{Time: at.Add(time.Minute)},
			{Time: at.Add(2 * time.Minute), ResourceUsage: logging.ResourceUsage{CPU: f(0.6), CPULimit: 2,
				Memory: f(80), MemoryLimit: 1024, OOMKills: 1}},
		}})

	assert.Len(t, data.Series, 3)
	assert.Nil(t, data.Series[1].CPU, "a step without a row is a gap")
	if assert.NotNil(t, data.Totals.CPU) {
		assert.InDelta(t, 0.4, *data.Totals.CPU, 1e-9)
	}
	assert.InDelta(t, 0.6, *data.Totals.CPUPeak, 1e-9)
	assert.InDelta(t, 100, *data.Totals.MemoryPeak, 1e-9)
	assert.InDelta(t, 2, data.Totals.CPULimit, 1e-9)
	assert.InDelta(t, 1024, data.Totals.MemoryLimit, 1e-9)
	assert.Equal(t, int64(1), data.Totals.OOMKills)
}
