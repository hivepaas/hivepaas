package taskdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A run a trigger fired says so: the event, the app it happened to, the
// deployment.
func TestTransformTaskSaysWhatTriggeredARun(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	task.MustSetArgs(&entity.TaskSchedJobExecArgs{Trigger: &entity.SchedJobTriggerCause{
		Event: base.SchedJobTriggerPostDeploy, AppID: "a1", DeploymentID: "d1",
	}})
	refObjects := entity.NewRefObjects()
	refObjects.RefApps["a1"] = &entity.App{ID: "a1", Name: "backend"}

	resp, err := TransformTask(task, nil, refObjects)

	assert.NoError(t, err)
	assert.Equal(t, &TaskTriggerResp{
		Event:        base.SchedJobTriggerPostDeploy,
		App:          &basedto.NamedObjectResp{ID: "a1", Name: "backend"},
		DeploymentID: "d1",
	}, resp.Trigger)
}

func TestTransformTaskGivesNoTriggerForARunByHand(t *testing.T) {
	resp, err := TransformTask(&entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	assert.Nil(t, resp.Trigger)
}
