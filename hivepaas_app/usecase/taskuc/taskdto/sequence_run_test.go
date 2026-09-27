package taskdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestTransformTaskGivesASequencesRun(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	task.MustSetOutput(&entity.SchedJobSeqRun{Started: true, CurrentStep: 1, Steps: []*entity.SchedJobSeqStepResult{
		{Job: entity.ObjectID{ID: "job-a"}, Status: base.SchedJobSeqStepDone},
		{Job: entity.ObjectID{ID: "job-b"}, Status: base.SchedJobSeqStepRunning},
	}})

	resp, err := TransformTask(task, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	if assert.NotNil(t, resp.SequenceRun) && assert.Len(t, resp.SequenceRun.Steps, 2) {
		assert.Equal(t, 1, resp.SequenceRun.CurrentStep)
		assert.Equal(t, base.SchedJobSeqStepRunning, resp.SequenceRun.Steps[1].Status)
	}
}

func TestTransformTaskGivesNoRunForAPlainJob(t *testing.T) {
	resp, err := TransformTask(&entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}, nil, entity.NewRefObjects())
	assert.NoError(t, err)
	assert.Nil(t, resp.SequenceRun)

	resp, err = TransformTask(&entity.Task{ID: "t2", Type: base.TaskTypeAppDeploy, Output: `{"x":1}`}, nil,
		entity.NewRefObjects())
	assert.NoError(t, err)
	assert.Nil(t, resp.SequenceRun)
}
