package queueimpl

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

var settleNow = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

func runningTask() *entity.Task {
	return &entity.Task{
		Status:    base.TaskStatusNotStarted,
		StartedAt: settleNow.Add(-time.Minute),
		Config:    entity.TaskConfig{MaxRetry: 2, RetryDelay: 60e9},
	}
}

func TestSettleTaskMarksASuccessDone(t *testing.T) {
	task := runningTask()
	rescheduleAt := settleTask(task, &queue.TaskExecData{Task: task}, nil, settleNow)

	assert.Equal(t, base.TaskStatusDone, task.Status)
	assert.True(t, rescheduleAt.IsZero())
}

func TestSettleTaskRunsAContinuedTaskAgainAtOnce(t *testing.T) {
	task := runningTask()
	task.RetryAt = settleNow.Add(-time.Hour)
	data := &queue.TaskExecData{Task: task}
	data.Continue()

	rescheduleAt := settleTask(task, data, nil, settleNow)

	assert.Equal(t, base.TaskStatusNotStarted, task.Status, "not done: it has more to do")
	assert.Equal(t, settleNow, task.RunAt)
	assert.True(t, task.RetryAt.IsZero())
	assert.Equal(t, settleNow, rescheduleAt)
}

func TestSettleTaskLetsACancelWinOverContinue(t *testing.T) {
	task := runningTask()
	data := &queue.TaskExecData{Task: task, TaskCanceled: true}
	data.Continue()

	rescheduleAt := settleTask(task, data, nil, settleNow)

	assert.Equal(t, base.TaskStatusCanceled, task.Status)
	assert.True(t, rescheduleAt.IsZero())
}

func TestSettleTaskRetriesAFailure(t *testing.T) {
	task := runningTask()
	data := &queue.TaskExecData{Task: task}
	data.Continue() // a failure is a failure, whatever was asked before it

	rescheduleAt := settleTask(task, data, errors.New("boom"), settleNow)

	assert.Equal(t, base.TaskStatusFailed, task.Status)
	assert.False(t, rescheduleAt.IsZero())
	assert.Equal(t, task.RetryAt, rescheduleAt)
}

func TestSettleTaskGivesUpOnANonRetryableFailure(t *testing.T) {
	task := runningTask()
	data := &queue.TaskExecData{Task: task, TaskNonRetryable: true}

	rescheduleAt := settleTask(task, data, errors.New("boom"), settleNow)

	assert.Equal(t, base.TaskStatusFailed, task.Status)
	assert.True(t, rescheduleAt.IsZero())
	assert.False(t, task.CanRetry())
}

// A task canceled while it runs is canceled, whatever error its work returned
// for being cut - a stream closed, a context done - and it is not retried.
func TestSettleTaskLetsACancelWinOverTheErrorItCaused(t *testing.T) {
	task := runningTask()
	data := &queue.TaskExecData{Task: task, TaskCanceled: true}

	rescheduleAt := settleTask(task, data, errors.New("stream closed without exit code"), settleNow)

	assert.Equal(t, base.TaskStatusCanceled, task.Status)
	assert.True(t, rescheduleAt.IsZero())
	assert.True(t, task.RetryAt.IsZero())
}
