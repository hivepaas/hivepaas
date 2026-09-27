package schedjobtriggerserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// An app enabled fires app-enabled, one disabled app-disabled; nothing else is
// a change of status.
func TestAppStatusEvent(t *testing.T) {
	event, ok := appStatusEvent(base.AppStatusActive)
	assert.True(t, ok)
	assert.Equal(t, base.SchedJobTriggerAppEnabled, event)

	event, ok = appStatusEvent(base.AppStatusDisabled)
	assert.True(t, ok)
	assert.Equal(t, base.SchedJobTriggerAppDisabled, event)

	_, ok = appStatusEvent(base.AppStatusDeleting)
	assert.False(t, ok)
}
