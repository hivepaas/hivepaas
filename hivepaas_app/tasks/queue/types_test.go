package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Only the queue calls these, and only on the data it made, so work registered
// by something running inside a task has to end up on that task's own data.
func TestSubTaskRegistersOnTheTaskItRunsInside(t *testing.T) {
	owner := &TaskExecData{Task: &entity.Task{ID: "owner"}}
	sub := owner.SubTask(&entity.Task{ID: "sub"})

	var ran []string
	sub.OnPostTx(func() { ran = append(ran, "post") })
	sub.OnEndTx(func() { ran = append(ran, "end") })
	sub.OnCommand(func(base.TaskCommand, ...any) { ran = append(ran, "command") })

	assert.Nil(t, sub.OnPostTxFunc)
	assert.Nil(t, sub.OnEndTxFunc)
	assert.Nil(t, sub.OnCommandFunc)

	owner.OnEndTxFunc()
	owner.OnPostTxFunc()
	owner.OnCommandFunc(base.TaskCommandCancel)

	assert.Equal(t, []string{"end", "post", "command"}, ran)
}

// The sub-task is the work's own - a workflow step, a preview's clone - and is
// told what the task around it already knows.
func TestSubTaskCarriesTheReferencesAndTheTaskItIsGiven(t *testing.T) {
	refObjects := entity.NewRefObjects()
	owner := &TaskExecData{Task: &entity.Task{ID: "owner"}, RefObjects: refObjects}

	sub := owner.SubTask(&entity.Task{ID: "sub"})

	assert.Equal(t, "sub", sub.Task.ID)
	assert.Same(t, refObjects, sub.RefObjects)
	assert.Same(t, owner.LogStore, sub.LogStore)
}

// Work done for a task that has none of its own - an agent, a request that is
// not a task - still gets data it can use, with nothing behind it.
func TestSubTaskOfNothing(t *testing.T) {
	var owner *TaskExecData

	sub := owner.SubTask(&entity.Task{ID: "sub"})

	assert.Equal(t, "sub", sub.Task.ID)
	sub.OnPostTx(func() {})
	assert.NotNil(t, sub.OnPostTxFunc)
}

// A sub-task's own sub-task is the same transaction still.
func TestSubTaskOfASubTaskReachesTheSameTask(t *testing.T) {
	owner := &TaskExecData{Task: &entity.Task{ID: "owner"}}

	deeper := owner.SubTask(&entity.Task{ID: "sub"}).SubTask(&entity.Task{ID: "deeper"})
	ran := false
	deeper.OnPostTx(func() { ran = true })

	owner.OnPostTxFunc()

	assert.True(t, ran)
}
