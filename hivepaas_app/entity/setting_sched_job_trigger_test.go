package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// An app's job listens to its own app's events; an env job to the apps its
// trigger names.
func TestSchedJobListensTo(t *testing.T) {
	appJob := &SchedJob{Triggers: []*SchedJobTrigger{
		{Event: base.SchedJobTriggerPostDeploy},
	}}
	listens, _ := appJob.ListensTo(base.SchedJobTriggerPostDeploy, "a1", true)
	assert.True(t, listens, "its own app")
	listens, _ = appJob.ListensTo(base.SchedJobTriggerPostDeploy, "a2", false)
	assert.False(t, listens, "another app")
	listens, _ = appJob.ListensTo(base.SchedJobTriggerDeployFailed, "a1", true)
	assert.False(t, listens, "another event")

	envJob := &SchedJob{Triggers: []*SchedJobTrigger{
		{Event: base.SchedJobTriggerHealthDown, Apps: []ObjectID{{ID: "a1"}, {ID: "a2"}}},
	}}
	listens, _ = envJob.ListensTo(base.SchedJobTriggerHealthDown, "a2", false)
	assert.True(t, listens)
	listens, _ = envJob.ListensTo(base.SchedJobTriggerHealthDown, "a3", false)
	assert.False(t, listens)

	var none *SchedJob
	listens, _ = none.ListensTo(base.SchedJobTriggerPostDeploy, "a1", true)
	assert.False(t, listens)
}

// A job listening to an event through a trigger that waits is waited for.
func TestSchedJobListensToSaysWhetherToWait(t *testing.T) {
	job := &SchedJob{Triggers: []*SchedJobTrigger{
		{Event: base.SchedJobTriggerPreDeploy, Apps: []ObjectID{{ID: "a1"}, {ID: "a2"}}},
		{Event: base.SchedJobTriggerPreDeploy, Apps: []ObjectID{{ID: "a1"}}, Wait: true},
	}}

	listens, wait := job.ListensTo(base.SchedJobTriggerPreDeploy, "a1", false)
	assert.True(t, listens)
	assert.True(t, wait)

	listens, wait = job.ListensTo(base.SchedJobTriggerPreDeploy, "a2", false)
	assert.True(t, listens)
	assert.False(t, wait)
}

// The apps a job's triggers name are its references.
func TestSchedJobTriggerAppsAreReferences(t *testing.T) {
	job := &SchedJob{JobType: base.SchedJobTypeJobSequence, Triggers: []*SchedJobTrigger{
		{Event: base.SchedJobTriggerPostDeploy, Apps: []ObjectID{{ID: "a1"}, {ID: "a2"}}},
		{Event: base.SchedJobTriggerHealthDown, Apps: []ObjectID{{ID: "a1"}}},
	}}

	assert.Equal(t, []string{"a1", "a2"}, job.GetRefObjectIDs().RefAppIDs)
}

// What fired a job's run is kept in its task's args.
func TestTaskArgsAsSchedJobExec(t *testing.T) {
	task := &Task{Type: base.TaskTypeSchedJobExec}
	args, err := task.ArgsAsSchedJobExec()
	assert.NoError(t, err)
	assert.Nil(t, args, "a run no trigger fired has no args")

	task.MustSetArgs(&TaskSchedJobExecArgs{Trigger: &SchedJobTriggerCause{
		Event: base.SchedJobTriggerPostDeploy, AppID: "a1", DeploymentID: "d1",
	}})
	reread := &Task{Type: base.TaskTypeSchedJobExec, Args: task.Args}
	args, err = reread.ArgsAsSchedJobExec()
	assert.NoError(t, err)
	if assert.NotNil(t, args) && assert.NotNil(t, args.Trigger) {
		assert.Equal(t, base.SchedJobTriggerPostDeploy, args.Trigger.Event)
		assert.Equal(t, "a1", args.Trigger.AppID)
		assert.Equal(t, "d1", args.Trigger.DeploymentID)
	}
}

// A run a trigger fired references the app the event happened to, so the run's
// page can name it.
func TestATriggeredRunReferencesItsApp(t *testing.T) {
	task := &Task{Type: base.TaskTypeSchedJobExec, Scope: base.ObjectScopeProjectEnv, ObjectID: "p1:dev"}
	task.MustSetArgs(&TaskSchedJobExecArgs{Trigger: &SchedJobTriggerCause{
		Event: base.SchedJobTriggerPostDeploy, AppID: "a1",
	}})

	refs := task.GetRefObjectIDs()

	assert.Equal(t, []string{"a1"}, refs.RefAppIDs)
	assert.Equal(t, []string{"p1:dev"}, refs.RefProjectEnvIDs)
}
