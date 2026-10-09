package queueimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A job turned off, or given another schedule, has its coming runs taken off.
// One that never started is removed: turned on again, the job makes it anew -
// the same job at the same time - where a canceled one would stand in the way.
// One that has started is canceled, and kept.
func TestUnschedulingRemovesTheRunsThatNeverStarted(t *testing.T) {
	now := time.Date(2026, 10, 9, 16, 30, 0, 0, time.UTC)
	coming := &entity.Task{Status: base.TaskStatusNotStarted, RunAt: now.Add(time.Minute)}
	running := &entity.Task{Status: base.TaskStatusInProgress, StartedAt: now.Add(-time.Minute)}
	goingOn := &entity.Task{Status: base.TaskStatusNotStarted, StartedAt: now.Add(-time.Minute)}
	done := &entity.Task{Status: base.TaskStatusDone, StartedAt: now.Add(-time.Hour)}

	unscheduled := unscheduleTasks([]*entity.Task{coming, running, goingOn, done}, now)

	assert.Equal(t, []*entity.Task{coming, running, goingOn}, unscheduled)
	assert.Equal(t, base.TaskStatusCanceled, coming.Status)
	assert.Equal(t, now, coming.DeletedAt, "removed: its time is free for a run made anew")
	assert.Equal(t, base.TaskStatusCanceled, running.Status)
	assert.True(t, running.DeletedAt.IsZero(), "a run that started is kept")
	assert.True(t, goingOn.DeletedAt.IsZero(), "a sequence's run between two steps is kept")
	assert.Equal(t, base.TaskStatusDone, done.Status)
}
