package hpappsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// spyTaskQueue counts the pauses asked of it: of this process alone, and of
// every process.
type spyTaskQueue struct {
	queue.TaskQueue
	local, all int
}

func (q *spyTaskQueue) StopScheduler() error {
	q.local++
	return nil
}

func (q *spyTaskQueue) StopAllSchedulers() error {
	q.all++
	return nil
}

func pauseFor(data *updateServiceSettingsData) (*spyTaskQueue, error) {
	spy := &spyTaskQueue{}
	uc := &UC{taskQueue: spy}
	return spy, uc.pauseTaskQueues(data)
}

// A change replacing only the worker pauses nothing. The request is handled by
// the app, which stays: paused, it would hold its tasks for nothing - on an
// install running the worker in the app, every task. Nor can the worker be told
// on its own: a control message goes to whichever process reads it first, the
// app as likely as the worker.
func TestAChangeToTheWorkerOnlyPausesNoTaskQueue(t *testing.T) {
	data := &updateServiceSettingsData{workerSvcChanges: true}

	spy, err := pauseFor(data)

	assert.NoError(t, err)
	assert.Equal(t, 0, spy.local)
	assert.Equal(t, 0, spy.all)
	assert.False(t, data.taskQueueStopped)
}

// A change replacing only the app pauses the process handling it, which goes
// with it, and not the worker, which stays.
func TestAChangeToTheAppOnlyPausesThisProcessOnly(t *testing.T) {
	data := &updateServiceSettingsData{mainSvcChanges: true}

	spy, err := pauseFor(data)

	assert.NoError(t, err)
	assert.Equal(t, 1, spy.local)
	assert.Equal(t, 0, spy.all)
	assert.True(t, data.taskQueueStopped)
}

// A change replacing both pauses every process it reaches: none stays.
func TestAChangeToBothPausesEveryTaskQueue(t *testing.T) {
	data := &updateServiceSettingsData{mainSvcChanges: true, workerSvcChanges: true}

	spy, err := pauseFor(data)

	assert.NoError(t, err)
	assert.Equal(t, 1, spy.all)
	assert.True(t, data.taskQueueStopped)
}
