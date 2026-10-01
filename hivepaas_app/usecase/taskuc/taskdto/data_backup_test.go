package taskdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A data backup's run gives the snapshot it took, and no sequence run: its
// output reads as an empty one.
func TestTransformTaskGivesADataBackupsSnapshot(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	task.MustSetOutput(&entity.SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42})

	resp, err := TransformTask(task, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	assert.Equal(t, &entity.SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42}, resp.DataBackup)
	assert.Nil(t, resp.SequenceRun)
}

func TestTransformTaskGivesNoSnapshotForASequencesRun(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	task.MustSetOutput(&entity.SchedJobSeqRun{Started: true})

	resp, err := TransformTask(task, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	assert.Nil(t, resp.DataBackup)
	assert.NotNil(t, resp.SequenceRun)
}

// A function call's run gives the response, and neither a snapshot nor a
// sequence run.
func TestTransformTaskGivesAFunctionCallsResponse(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	result := &entity.SchedJobFunctionInvokeResult{Outcome: "ok", Status: 404, Body: []byte("no")}
	task.MustSetOutput(result)

	resp, err := TransformTask(task, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	assert.Equal(t, result, resp.FunctionInvoke)
	assert.Nil(t, resp.DataBackup)
	assert.Nil(t, resp.SequenceRun)
}
