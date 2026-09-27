package schedjobuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The apps a job's triggers name are checked by checkTriggerApps, which knows
// an env's apps; the setting base would look for them in the job's own scope.
func TestTriggerAppsAreCheckedApartFromTheJobsOtherReferences(t *testing.T) {
	job := &entity.SchedJob{
		JobType: base.SchedJobTypeJobSequence,
		Triggers: []*entity.SchedJobTrigger{
			{Event: base.SchedJobTriggerPostDeploy, Apps: []entity.ObjectID{{ID: "a1"}}},
		},
	}

	assert.Equal(t, []string{"a1"}, job.GetRefObjectIDs().RefAppIDs)
	assert.Empty(t, verifyingRefIDs(job).RefAppIDs)
}

func TestTriggerAppProblem(t *testing.T) {
	appScope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "a1"}
	envScope := &entity.ObjectScope{ScopeType: base.ObjectScopeProjectEnv, ProjectEnvID: "p1:dev"}
	appOf := func(env string) *entity.App { return &entity.App{ID: "a2", ProjectEnvID: env} }
	postDeploy := func(appIDs ...string) *entity.SchedJobTrigger {
		trigger := &entity.SchedJobTrigger{Event: base.SchedJobTriggerPostDeploy}
		for _, id := range appIDs {
			trigger.Apps = append(trigger.Apps, entity.ObjectID{ID: id})
		}
		return trigger
	}

	assert.Empty(t, triggerAppProblem(appScope, postDeploy(), nil))
	assert.Equal(t, "an app's job listens to its own app: name no app",
		triggerAppProblem(appScope, postDeploy("a2"), map[string]*entity.App{"a2": appOf("p1:dev")}))

	apps := map[string]*entity.App{"a2": appOf("p1:dev"), "a3": appOf("p1:prod")}
	assert.Empty(t, triggerAppProblem(envScope, postDeploy("a2"), apps))
	assert.Equal(t, "name the apps whose events run the job", triggerAppProblem(envScope, postDeploy(), apps))
	assert.Equal(t, "app a3 is not in this env", triggerAppProblem(envScope, postDeploy("a2", "a3"), apps))
	assert.Equal(t, "app a4 is not found", triggerAppProblem(envScope, postDeploy("a4"), apps))
}
