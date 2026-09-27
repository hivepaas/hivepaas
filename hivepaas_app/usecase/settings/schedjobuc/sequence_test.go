package schedjobuc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestJobSequencesLiveInAppsAndEnvs(t *testing.T) {
	seq := base.SchedJobTypeJobSequence
	cmd := base.SchedJobTypeContainerCommand

	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeApp, seq))
	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeProjectEnv, seq))
	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeApp, cmd))
	assert.NoError(t, checkJobTypeInScope(base.ObjectScopeGlobal, base.SchedJobTypeSystemCleanup))

	for _, scope := range []base.ObjectScopeType{base.ObjectScopeGlobal, base.ObjectScopeProject} {
		err := checkJobTypeInScope(scope, seq)
		assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "%s: got %v", scope, err)
	}
	err := checkJobTypeInScope(base.ObjectScopeProjectEnv, cmd)
	assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "an env holds sequences only, for now: got %v", err)
}

func TestASequencesMembersAreCheckedApartFromItsOtherReferences(t *testing.T) {
	job := &entity.SchedJob{
		JobType:      base.SchedJobTypeJobSequence,
		Notification: nil,
		Sequence: &entity.SchedJobSequence{Steps: []*entity.SchedJobSequenceStep{
			{Job: entity.ObjectID{ID: "job-a"}}, {Job: entity.ObjectID{ID: "job-b"}},
		}},
	}

	assert.ElementsMatch(t, []string{"job-a", "job-b"}, job.GetRefObjectIDs().RefSettingIDs)
	assert.Empty(t, verifyingRefIDs(job).RefSettingIDs,
		"members are checked by checkSequenceMembers, which can see an env's apps and a disabled job")
}

func TestMemberProblem(t *testing.T) {
	appScope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app-1"}
	envScope := &entity.ObjectScope{ScopeType: base.ObjectScopeProjectEnv, ProjectEnvID: "env-1"}
	appJob := func(appID string, jobType base.SchedJobType) *entity.Setting {
		return &entity.Setting{ID: "job", Scope: base.ObjectScopeApp, ObjectID: appID, Kind: string(jobType),
			Status: base.SettingStatusDisabled}
	}
	inEnv := &entity.App{ID: "app-2", ProjectEnvID: "env-1"}
	elsewhere := &entity.App{ID: "app-3", ProjectEnvID: "env-2"}
	cmd := base.SchedJobTypeContainerCommand

	assert.Empty(t, memberProblem(appScope, appJob("app-1", cmd), nil), "the app's own job, even disabled")
	assert.Empty(t, memberProblem(envScope, appJob("app-2", cmd), inEnv), "a job of an app in the env")

	assert.Contains(t, memberProblem(appScope, nil, nil), "not found")
	assert.Contains(t, memberProblem(appScope, appJob("app-9", cmd), nil), "another app")
	assert.Contains(t, memberProblem(envScope, appJob("app-3", cmd), elsewhere), "outside this env")
	assert.Contains(t, memberProblem(envScope, appJob("app-4", cmd), nil), "outside this env")
	assert.Contains(t, memberProblem(appScope, appJob("app-1", base.SchedJobTypeJobSequence), nil),
		"another sequence")
	envJob := &entity.Setting{ID: "job", Scope: base.ObjectScopeProjectEnv, ObjectID: "env-1", Kind: string(cmd)}
	assert.Contains(t, memberProblem(envScope, envJob, nil), "not an app's job")
}
