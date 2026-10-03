package appsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
)

// An app is paused only when none of the signals it scales on can be read;
// one that can is enough.
func TestUnreadableSignals(t *testing.T) {
	both := &entity.AppAutoscale{RequestsTarget: 10, CPUTarget: 70}
	onCPU := &entity.AppAutoscale{CPUTarget: 70}

	assert.Empty(t, unreadableSignals(both, &appautoscaleservice.Check{Requests: "not-exposed"}))
	assert.Empty(t, unreadableSignals(both, &appautoscaleservice.Check{CPU: "no-cpu-limit"}))
	assert.Equal(t, "not-exposed",
		unreadableSignals(both, &appautoscaleservice.Check{Requests: "not-exposed", CPU: "no-cpu-limit"}))
	assert.Equal(t, "no-cpu-limit", unreadableSignals(onCPU, &appautoscaleservice.Check{CPU: "no-cpu-limit"}))
	assert.Empty(t, unreadableSignals(onCPU, &appautoscaleservice.Check{Requests: "not-exposed"}),
		"requests it does not scale on")
	assert.Empty(t, unreadableSignals(&entity.AppAutoscale{}, &appautoscaleservice.Check{CPU: "x"}))
}
