package schedjobexecserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A run a trigger fired is told which event, of which app, and which deployment.
func TestTriggerEnv(t *testing.T) {
	task := &entity.Task{Type: base.TaskTypeSchedJobExec}
	assert.Empty(t, triggerEnv(task), "a run no trigger fired is told nothing")

	task.MustSetArgs(&entity.TaskSchedJobExecArgs{Trigger: &entity.SchedJobTriggerCause{
		Event: base.SchedJobTriggerPostDeploy, AppID: "a1", DeploymentID: "d1",
	}})
	assert.Equal(t, []string{
		"HIVEPAAS_TRIGGER_EVENT=post-deploy",
		"HIVEPAAS_TRIGGER_APP=a1",
		"HIVEPAAS_TRIGGER_DEPLOYMENT=d1",
	}, triggerEnv(task))
}
