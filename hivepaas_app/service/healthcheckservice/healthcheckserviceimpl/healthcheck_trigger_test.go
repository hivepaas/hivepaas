package healthcheckserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
)

// A check fires health-down or health-up only when its state changed from one it
// knew: not on its first run, nor when its last state was lost.
func TestHealthEvent(t *testing.T) {
	up := &cacheentity.HealthcheckState{State: base.HealthcheckStateSuccess}
	down := &cacheentity.HealthcheckState{State: base.HealthcheckStateFailure}

	event, ok := healthEvent(up, base.TaskStatusFailed)
	assert.True(t, ok)
	assert.Equal(t, base.SchedJobTriggerHealthDown, event)

	event, ok = healthEvent(down, base.TaskStatusDone)
	assert.True(t, ok)
	assert.Equal(t, base.SchedJobTriggerHealthUp, event)

	_, ok = healthEvent(up, base.TaskStatusDone)
	assert.False(t, ok, "still up")
	_, ok = healthEvent(down, base.TaskStatusFailed)
	assert.False(t, ok, "still down")
	_, ok = healthEvent(nil, base.TaskStatusDone)
	assert.False(t, ok, "no state known before")
}
