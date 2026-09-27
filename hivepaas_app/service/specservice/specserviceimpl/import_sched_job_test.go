package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const devSettingsPath = "projects/project_a/envs/dev/settings"

// An env's job sequence is imported: its step names the app's job by the
// installation's id, and its task is scheduled as the dashboard schedules one.
func TestAnEnvSequenceIsImportedWithItsStepsMapped(t *testing.T) {
	svc, bundle := planFixture(t)
	release := bundle.Envs["project_a"]["dev"].Settings["schedJobs"].(map[string]any)["release"].(map[string]any)
	release["sequence"].(map[string]any)["onFailure"] = "continue"

	env := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), devSettingsPath)
	assert.Equal(t, []string{"schedJobs/release"}, env.Changes)
	assert.Empty(t, issuesOf(env, specmodel.CodeTypeNotImportable))

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	written := persistedSetting(t, svc, func(s *entity.Setting) bool {
		return s.Type == base.SettingTypeSchedJob && s.Scope == base.ObjectScopeProjectEnv
	})
	job := written.MustAsSchedJob()
	assert.Equal(t, base.SchedJobSeqOnFailureContinue, job.Sequence.OnFailure)
	if assert.Len(t, job.Sequence.Steps, 1) {
		assert.Equal(t, "job_1", job.Sequence.Steps[0].Job.ID)
	}
	scheduled := svc.taskQueue.(*fakeTaskQueue).scheduledJobs
	if assert.Len(t, scheduled, 1) {
		assert.Equal(t, written.ID, scheduled[0].ID)
	}
}

// A step naming a job the import does not find is reported, and the sequence
// is written pending, its step empty: the run skips it.
func TestAnEnvSequenceWithAStepNotFoundIsWrittenPending(t *testing.T) {
	svc, bundle := planFixture(t)
	release := bundle.Envs["project_a"]["dev"].Settings["schedJobs"].(map[string]any)["release"].(map[string]any)
	step := release["sequence"].(map[string]any)["steps"].([]any)[0].(map[string]any)
	step["job"] = map[string]any{"id": backendPath + "/schedJobs/gone"}

	env := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), devSettingsPath)
	if issues := issuesOf(env, specmodel.CodeRefNotFound); assert.Len(t, issues, 1) {
		assert.Equal(t, "schedJobs/release", issues[0].Detail["setting"])
	}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	written := persistedSetting(t, svc, func(s *entity.Setting) bool {
		return s.Type == base.SettingTypeSchedJob && s.Scope == base.ObjectScopeProjectEnv
	})
	assert.Equal(t, base.SettingStatusPending, written.Status)
	assert.Empty(t, written.MustAsSchedJob().Sequence.Steps[0].Job.ID)
}

// A project's scheduled job is still not imported: only an env's is.
func TestAProjectSchedJobIsSkipped(t *testing.T) {
	svc, bundle := planFixture(t)
	release := bundle.Envs["project_a"]["dev"].Settings["schedJobs"].(map[string]any)["release"]
	bundle.Projects["project_a"].Settings["schedJobs"] = map[string]any{"release": release}

	project := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), "projects/project_a/settings")

	assert.NotContains(t, project.Changes, "schedJobs/release")
	if issues := issuesOf(project, specmodel.CodeTypeNotImportable); assert.Len(t, issues, 1) {
		assert.Equal(t, reasonSchedulesTasks, issues[0].Detail["reason"])
	}
}

func TestImportPolicyInAScope(t *testing.T) {
	assert.Empty(t, importPolicyIn(base.SettingTypeSchedJob, base.ObjectScopeProjectEnv).skip)
	assert.Equal(t, reasonSchedulesTasks, importPolicyIn(base.SettingTypeSchedJob, base.ObjectScopeProject).skip)
	assert.Equal(t, reasonSchedulesTasks, importPolicyIn(base.SettingTypeSchedJob, base.ObjectScopeGlobal).skip)
	assert.Equal(t, reasonSchedulesTasks, importPolicyIn(base.SettingTypePeriodicJob, base.ObjectScopeProjectEnv).skip)
}

// A trigger's apps travel as the IDs they had where the bundle was made, and are
// written as the IDs the same apps have here.
func TestAnEnvJobsTriggerAppsAreMappedOnImport(t *testing.T) {
	svc, bundle := planFixture(t)
	backend := bundle.Envs["project_a"]["dev"].Apps["backend"]
	backend.ID = "app_elsewhere"
	release := bundle.Envs["project_a"]["dev"].Settings["schedJobs"].(map[string]any)["release"].(map[string]any)
	triggers := release["triggers"].([]any)
	assert.Equal(t, "app_1", triggers[0].(map[string]any)["apps"].([]any)[0].(map[string]any)["id"],
		"export writes the app's ID")
	triggers[0].(map[string]any)["apps"] = []any{map[string]any{"id": "app_elsewhere"}}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	written := persistedSetting(t, svc, func(s *entity.Setting) bool {
		return s.Type == base.SettingTypeSchedJob && s.Scope == base.ObjectScopeProjectEnv
	})
	job := written.MustAsSchedJob()
	if assert.Len(t, job.Triggers, 1) {
		assert.Equal(t, []entity.ObjectID{{ID: "app_1"}}, job.Triggers[0].Apps)
	}
}
