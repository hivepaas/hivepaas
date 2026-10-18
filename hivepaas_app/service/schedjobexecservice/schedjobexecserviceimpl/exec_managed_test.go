package schedjobexecserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// A managed job writes outputs as a sequence's step does, to a file of its own
// run: what its engine reads its state from. Any other job has none.
func TestAManagedJobHasAnOutputFile(t *testing.T) {
	managed := &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand, ManagedBy: entity.ObjectID{ID: "m1"}}
	ordinary := &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand}
	run := &queue.TaskExecData{Task: &entity.Task{ID: "t1"}}

	file, env := outputFileFor(&schedjobexecservice.SchedJobExecReq{TaskExecData: run}, managed)
	assert.Equal(t, schedjobexecservice.OutputFilePath("t1", 0), file)
	assert.Equal(t, []string{"HIVEPAAS_OUTPUT=" + file}, env)

	file, env = outputFileFor(&schedjobexecservice.SchedJobExecReq{TaskExecData: run}, ordinary)
	assert.Empty(t, file)
	assert.Empty(t, env)

	step := &schedjobexecservice.SequenceStep{Step: 2, Steps: 3, OutputFile: "/tmp/step"}
	file, env = outputFileFor(&schedjobexecservice.SchedJobExecReq{TaskExecData: run, Sequence: step}, managed)
	assert.Equal(t, "/tmp/step", file, "a step's is the sequence's")
	assert.Empty(t, env, "sequenceEnv already says where")
}
