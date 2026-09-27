package schedjobdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestTransformEnvSchedJobsSaysWhoseEachJobIs(t *testing.T) {
	envSeq := &entity.Setting{ID: "seq", Type: base.SettingTypeSchedJob, Name: "release",
		Scope: base.ObjectScopeProjectEnv, ObjectID: "env-1", Kind: string(base.SchedJobTypeJobSequence)}
	envSeq.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeJobSequence,
		Sequence: &entity.SchedJobSequence{Steps: []*entity.SchedJobSequenceStep{{Job: entity.ObjectID{ID: "cmd"}}}}})
	appCmd := &entity.Setting{ID: "cmd", Type: base.SettingTypeSchedJob, Name: "migrate",
		Scope: base.ObjectScopeApp, ObjectID: "app-1", Kind: string(base.SchedJobTypeContainerCommand)}
	appCmd.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeContainerCommand})
	refObjects := entity.NewRefObjects()
	refObjects.RefApps["app-1"] = &entity.App{ID: "app-1", Name: "backend"}

	resp, err := TransformEnvSchedJobs([]*entity.Setting{envSeq, appCmd}, refObjects)

	assert.NoError(t, err)
	if assert.Len(t, resp, 2) {
		assert.Equal(t, base.ObjectScopeProjectEnv, resp[0].Scope)
		assert.Nil(t, resp[0].OwnerApp)
		assert.Equal(t, "release", resp[0].Name)
		assert.Equal(t, base.ObjectScopeApp, resp[1].Scope)
		if assert.NotNil(t, resp[1].OwnerApp) {
			assert.Equal(t, "backend", resp[1].OwnerApp.Name)
		}
	}
}

func TestListEnvSchedJobReqFiltersByJobTypeAndApp(t *testing.T) {
	req := NewListEnvSchedJobReq()
	req.JobTypes = []base.SchedJobType{base.SchedJobTypeContainerCommand}
	req.AppID = jobA
	assert.Empty(t, req.Validate())

	req.JobTypes = []base.SchedJobType{"nope"}
	assert.NotEmpty(t, req.Validate())
}
