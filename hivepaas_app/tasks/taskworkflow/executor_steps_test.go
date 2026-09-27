package taskworkflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// stepQueue runs a step by noting its type; the embedded interface leaves the
// rest nil, which a workflow step never reaches.
type stepQueue struct {
	queue.TaskQueue
	ran []base.TaskType
}

func (q *stepQueue) ExecuteTaskType(_ context.Context, _ database.Tx, typ base.TaskType, _ *queue.TaskExecData) error {
	q.ran = append(q.ran, typ)
	return nil
}

func TestWorkflowRunsOneStepPerExecutionAndContinues(t *testing.T) {
	q := &stepQueue{}
	e := &Executor{taskQueue: q}
	task := &entity.Task{ID: "wf", Type: base.TaskTypeWorkflow}
	assert.NoError(t, task.SetArgs(&entity.WorkflowArgs{Steps: []*entity.WorkflowStep{
		{Name: "one", Type: base.TaskTypeDummy},
		{Name: "two", Type: base.TaskTypeAppDeploy},
	}}))

	first := &queue.TaskExecData{Task: task}
	assert.NoError(t, e.execute(context.Background(), database.Tx{}, first))
	assert.True(t, first.Continued(), "a step is left: run again")

	second := &queue.TaskExecData{Task: task}
	assert.NoError(t, e.execute(context.Background(), database.Tx{}, second))
	assert.False(t, second.Continued(), "the last step: done")

	assert.Equal(t, []base.TaskType{base.TaskTypeDummy, base.TaskTypeAppDeploy}, q.ran)
}
