package schedjobdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const appA = "01JAB9XED0GTXBSQDFVYAJ8WA9"

func triggered(triggers ...*SchedJobTriggerReq) *CreateSchedJobReq {
	req := sequenceReq(jobA)
	req.Triggers = triggers
	return req
}

func TestTriggersAreKeptOnTheJob(t *testing.T) {
	req := triggered(
		&SchedJobTriggerReq{Event: base.SchedJobTriggerPreDeploy, Wait: true},
		&SchedJobTriggerReq{Event: base.SchedJobTriggerHealthDown, Apps: []basedto.ObjectIDReq{{ID: appA}}},
	)

	assert.Equal(t, "", invalidFields(t, req))
	job := req.ToEntity()
	if assert.Len(t, job.Triggers, 2) {
		assert.Equal(t, base.SchedJobTriggerPreDeploy, job.Triggers[0].Event)
		assert.True(t, job.Triggers[0].Wait)
		assert.Equal(t, []entity.ObjectID{{ID: appA}}, job.Triggers[1].Apps)
	}
}

func TestATriggerNeedsAKnownEvent(t *testing.T) {
	assert.Equal(t, "triggers[0].event", invalidFields(t, triggered(&SchedJobTriggerReq{Event: "pushed"})))
}

func TestOnlyAPreDeployTriggerWaits(t *testing.T) {
	req := triggered(&SchedJobTriggerReq{Event: base.SchedJobTriggerPostDeploy, Wait: true})
	assert.Equal(t, "triggers[0].wait", invalidFields(t, req))
}

func TestAJobHasAtMostTenTriggersNoTwoAlike(t *testing.T) {
	var many []*SchedJobTriggerReq
	for range base.SchedJobMaxTriggers + 1 {
		many = append(many, &SchedJobTriggerReq{Event: base.SchedJobTriggerPostDeploy})
	}
	assert.Contains(t, invalidFields(t, triggered(many...)), "triggers ")

	req := triggered(
		&SchedJobTriggerReq{Event: base.SchedJobTriggerHealthUp, Apps: []basedto.ObjectIDReq{{ID: appA}}},
		&SchedJobTriggerReq{Event: base.SchedJobTriggerHealthUp, Apps: []basedto.ObjectIDReq{{ID: appA}}},
	)
	assert.Equal(t, "triggers[1]", invalidFields(t, req))
}

func TestASystemJobHasNoTriggers(t *testing.T) {
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{
		Name:     "renew",
		JobType:  base.SchedJobTypeSSLRenewal,
		Triggers: []*SchedJobTriggerReq{{Event: base.SchedJobTriggerPostDeploy}},
	}
	assert.Contains(t, invalidFields(t, req), "triggers")
}

func TestTransformSchedJobTriggersNamesTheApps(t *testing.T) {
	refObjects := entity.NewRefObjects()
	refObjects.RefApps["a1"] = &entity.App{ID: "a1", Name: "backend"}

	resp := TransformSchedJobTriggers([]*entity.SchedJobTrigger{
		{Event: base.SchedJobTriggerPostDeploy, Apps: []entity.ObjectID{{ID: "a1"}, {ID: "gone"}}},
	}, refObjects)

	if assert.Len(t, resp, 1) && assert.Len(t, resp[0].Apps, 2) {
		assert.Equal(t, base.SchedJobTriggerPostDeploy, resp[0].Event)
		assert.Equal(t, &basedto.NamedObjectResp{ID: "a1", Name: "backend"}, resp[0].Apps[0])
		assert.Equal(t, "gone", resp[0].Apps[1].ID)
	}
	assert.Empty(t, TransformSchedJobTriggers(nil, refObjects))
}
