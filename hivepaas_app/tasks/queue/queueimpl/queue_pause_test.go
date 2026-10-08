package queueimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// A process that runs no task queue has nothing to pause: the app, when its
// tasks run in the worker. That is no failure - a change to the app's settings
// pauses the process handling it, whichever it is.
func TestPausingAProcessWithoutATaskQueueDoesNothing(t *testing.T) {
	q := &taskQueue{logger: logging.GlobalLogger()}

	assert.NoError(t, q.StopScheduler())
}
