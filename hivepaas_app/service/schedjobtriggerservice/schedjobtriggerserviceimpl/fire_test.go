package schedjobtriggerserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice/schedjobserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

var app1 = &entity.App{ID: "a1", Name: "backend", ProjectEnvID: "p1:dev"}

func jobSetting(t *testing.T, id string, scope base.ObjectScopeType, objectID string,
	triggers ...*entity.SchedJobTrigger,
) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Name: "job " + id, Type: base.SettingTypeSchedJob, Scope: scope,
		ObjectID: objectID, Status: base.SettingStatusActive, Kind: string(base.SchedJobTypeContainerCommand)}
	setting.MustSetData(&entity.SchedJob{
		JobType: base.SchedJobTypeContainerCommand, Timeout: 90_000_000_000, Triggers: triggers,
	})
	return setting
}

func onEvent(event base.SchedJobTriggerEvent, wait bool, appIDs ...string) *entity.SchedJobTrigger {
	trigger := &entity.SchedJobTrigger{Event: event, Wait: wait}
	for _, id := range appIDs {
		trigger.Apps = append(trigger.Apps, entity.ObjectID{ID: id})
	}
	return trigger
}

func newTestService() *service {
	return &service{schedJobService: schedjobserviceimpl.New(nil)}
}

// An event runs the app's own jobs listening to it, and the env's jobs naming
// the app; nothing else.
func TestBuildRunsPicksTheJobsListening(t *testing.T) {
	jobs := []*entity.Setting{
		jobSetting(t, "own", base.ObjectScopeApp, "a1", onEvent(base.SchedJobTriggerPostDeploy, false)),
		jobSetting(t, "env", base.ObjectScopeProjectEnv, "p1:dev",
			onEvent(base.SchedJobTriggerPostDeploy, false, "a1", "a2")),
		jobSetting(t, "env-other-app", base.ObjectScopeProjectEnv, "p1:dev",
			onEvent(base.SchedJobTriggerPostDeploy, false, "a2")),
		jobSetting(t, "other-event", base.ObjectScopeApp, "a1", onEvent(base.SchedJobTriggerDeployFailed, false)),
		jobSetting(t, "no-trigger", base.ObjectScopeApp, "a1"),
	}
	disabled := jobSetting(t, "disabled", base.ObjectScopeApp, "a1", onEvent(base.SchedJobTriggerPostDeploy, false))
	disabled.Status = base.SettingStatusDisabled
	jobs = append(jobs, disabled)

	runs, err := newTestService().buildRuns(app1, jobs, base.SchedJobTriggerPostDeploy,
		&schedjobtriggerservice.TriggerInfo{DeploymentID: "d1"}, true, time.Now())

	assert.NoError(t, err)
	var targets []string
	for _, run := range runs {
		targets = append(targets, run.Task.TargetID)
	}
	assert.Equal(t, []string{"own", "env"}, targets)
}

// A run's task says what fired it, and is a run as run-now makes one.
func TestBuildRunsRecordsTheCause(t *testing.T) {
	jobs := []*entity.Setting{
		jobSetting(t, "own", base.ObjectScopeApp, "a1", onEvent(base.SchedJobTriggerPreDeploy, true)),
	}

	runs, err := newTestService().buildRuns(app1, jobs, base.SchedJobTriggerPreDeploy,
		&schedjobtriggerservice.TriggerInfo{DeploymentID: "d1"}, true, time.Now())

	assert.NoError(t, err)
	if assert.Len(t, runs, 1) {
		run := runs[0]
		assert.True(t, run.Wait)
		assert.Equal(t, "job own", run.JobName)
		assert.Equal(t, 90*time.Second, run.Timeout)
		assert.Equal(t, base.TaskTypeSchedJobExec, run.Task.Type)
		assert.Equal(t, base.TaskStatusNotStarted, run.Task.Status)
		args, err := run.Task.ArgsAsSchedJobExec()
		assert.NoError(t, err)
		assert.Equal(t, &entity.SchedJobTriggerCause{
			Event: base.SchedJobTriggerPreDeploy, AppID: "a1", DeploymentID: "d1",
		}, args.Trigger)
	}
}

// An app with its scheduled jobs feature off runs none of its own jobs by
// trigger; the env's jobs that name it still run.
func TestBuildRunsLeavesOutTheAppsJobsWhenTheFeatureIsOff(t *testing.T) {
	jobs := []*entity.Setting{
		jobSetting(t, "own", base.ObjectScopeApp, "a1", onEvent(base.SchedJobTriggerHealthDown, false)),
		jobSetting(t, "env", base.ObjectScopeProjectEnv, "p1:dev", onEvent(base.SchedJobTriggerHealthDown, false, "a1")),
	}

	runs, err := newTestService().buildRuns(app1, jobs, base.SchedJobTriggerHealthDown, nil, false, time.Now())

	assert.NoError(t, err)
	if assert.Len(t, runs, 1) {
		assert.Equal(t, "env", runs[0].Task.TargetID)
	}
}
