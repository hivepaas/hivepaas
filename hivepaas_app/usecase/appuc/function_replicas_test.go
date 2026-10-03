package appuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
)

// The replicas at a step's end: before the first scaling its from, after one
// its to, now's count for the last step past the last scaling.
func TestReplicasAt(t *testing.T) {
	at := func(m int) time.Time { return time.Date(2026, 10, 3, 10, m, 0, 0, time.UTC) }
	events := []*appautoscaleservice.Event{
		{Time: at(5), From: 1, To: 2},
		{Time: at(20), From: 2, To: 4},
	}
	assert.Equal(t, 1, replicasAt(at(0), events, 3, false), "before the first")
	assert.Equal(t, 2, replicasAt(at(10), events, 3, false), "between")
	assert.Equal(t, 2, replicasAt(at(20).Add(-time.Second), events, 3, false))
	assert.Equal(t, 4, replicasAt(at(30), events, 3, false), "after the last")
	assert.Equal(t, 3, replicasAt(at(30), events, 3, true), "now's, changed by hand since")
	assert.Equal(t, 3, replicasAt(at(30), nil, 3, false), "no scaling: now's")
}
