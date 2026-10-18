package taskschedjobexec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A managed job's outputs are its run's output, failed or not: a health check
// that fails says why in them.
func TestAManagedJobsOutputsAreItsRunsOutput(t *testing.T) {
	task := &entity.Task{ID: "t1"}
	job := &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand, ManagedBy: entity.ObjectID{ID: "m1"}}

	keepManagedOutputs(task, job, &jobResult{outputs: map[string]string{"FAILED_COUNT": "2"}})

	output, err := task.OutputAsSchedJobOutputs()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"FAILED_COUNT": "2"}, output.Outputs)
}

func TestOtherJobsOutputsAreLeftAlone(t *testing.T) {
	for name, tc := range map[string]struct {
		job    *entity.SchedJob
		result *jobResult
	}{
		"an ordinary job": {&entity.SchedJob{JobType: base.SchedJobTypeContainerCommand},
			&jobResult{outputs: map[string]string{"A": "1"}}},
		"no outputs": {&entity.SchedJob{ManagedBy: entity.ObjectID{ID: "m1"}}, &jobResult{}},
		"no result":  {&entity.SchedJob{ManagedBy: entity.ObjectID{ID: "m1"}}, nil},
	} {
		task := &entity.Task{ID: "t1"}
		keepManagedOutputs(task, tc.job, tc.result)
		assert.Empty(t, task.Output, name)
	}
}

// A managed data backup's run keeps the snapshot it took beside its outputs: its
// run's page links the snapshot, and its engine reads the outputs.
func TestAManagedDataBackupKeepsItsSnapshot(t *testing.T) {
	task := &entity.Task{ID: "t1"}
	result := &entity.SchedJobDataBackupResult{SnapshotID: "k1a2b3", SizeBytes: 42}
	task.MustSetOutput(result)
	job := &entity.SchedJob{JobType: base.SchedJobTypeDataBackup, ManagedBy: entity.ObjectID{ID: "m1"}}

	keepManagedOutputs(task, job, &jobResult{outputs: dataBackupOutputs(result)})

	output, err := task.OutputAsSchedJobOutputs()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"SNAPSHOT_ID": "k1a2b3", "SNAPSHOT_SIZE_BYTES": "42"}, output.Outputs)
	snapshot, err := task.OutputAsDataBackup()
	assert.NoError(t, err)
	assert.Equal(t, result, snapshot)
}
