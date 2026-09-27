package schedjobdto

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	jobA = "01JAB9XED0GTXBSQDFVYAJ8WA7"
	jobB = "01JAB9XED0GTXBSQDFVYAJ8WA8"
)

func sequenceReq(jobIDs ...string) *CreateSchedJobReq {
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{
		Name:     "release",
		JobType:  base.SchedJobTypeJobSequence,
		Sequence: &SchedJobSequenceReq{},
	}
	for _, id := range jobIDs {
		req.Sequence.Steps = append(req.Sequence.Steps, &SchedJobSequenceStepReq{Job: basedto.ObjectIDReq{ID: id}})
	}
	return req
}

// invalidFields is the paths of the fields the request fails on, joined.
func invalidFields(t *testing.T, req *CreateSchedJobReq) string {
	t.Helper()
	assert.NoError(t, req.ModifyRequest())
	errs := req.Validate()
	if len(errs) == 0 {
		return ""
	}
	paths := make([]string, 0, len(errs))
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return strings.Join(paths, " ")
}

func TestASequenceIsValidWithoutASchedule(t *testing.T) {
	req := sequenceReq(jobA, jobB, jobA)

	assert.Equal(t, "", invalidFields(t, req))
	assert.Equal(t, base.SchedJobSeqModeSequential, req.Sequence.Mode, "the mode defaults to sequential")
	assert.Equal(t, base.SchedJobSeqOnFailureStop, req.Sequence.OnFailure, "on failure defaults to stop")

	job := req.ToEntity()
	assert.Nil(t, job.Schedule)
	if assert.NotNil(t, job.Sequence) && assert.Len(t, job.Sequence.Steps, 3) {
		assert.Equal(t, jobB, job.Sequence.Steps[1].Job.ID)
	}
}

func TestASequenceNeedsBetweenOneAndFiftySteps(t *testing.T) {
	assert.Contains(t, invalidFields(t, sequenceReq()), "sequence.steps")

	ids := make([]string, base.SchedJobSeqMaxSteps+1)
	for i := range ids {
		ids[i] = jobA
	}
	assert.Contains(t, invalidFields(t, sequenceReq(ids...)), "sequence.steps")
	assert.Equal(t, "", invalidFields(t, sequenceReq(ids[:base.SchedJobSeqMaxSteps]...)))
}

func TestASequenceStepNamesAJob(t *testing.T) {
	assert.Contains(t, invalidFields(t, sequenceReq("")), "sequence.steps[0].job")
}

func TestASequenceRefusesAnUnknownModeOrFailurePolicy(t *testing.T) {
	req := sequenceReq(jobA)
	req.Sequence.Mode = "parallel"
	assert.Contains(t, invalidFields(t, req), "sequence.mode")

	req = sequenceReq(jobA)
	req.Sequence.OnFailure = "retry"
	assert.Contains(t, invalidFields(t, req), "sequence.onFailure")
}

func TestASequenceRunsNoCommandOfItsOwn(t *testing.T) {
	req := sequenceReq(jobA)
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "echo hi"}
	assert.Contains(t, invalidFields(t, req), "command")

	req = sequenceReq(jobA)
	req.App = basedto.ObjectIDReq{ID: jobB}
	assert.Contains(t, invalidFields(t, req), "app")
}

func TestOnlyASequenceHasSteps(t *testing.T) {
	req := sequenceReq(jobA)
	req.JobType = base.SchedJobTypeSystemCleanup
	req.Schedule = &ScheduleReq{Interval: 3600e9, InitialTime: time.Now()}
	assert.Contains(t, invalidFields(t, req), "sequence")
}

func TestAnyJobMayGoWithoutASchedule(t *testing.T) {
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{Name: "cleanup", JobType: base.SchedJobTypeSystemCleanup}
	assert.Equal(t, "", invalidFields(t, req))
	assert.Nil(t, req.ToEntity().Schedule)
}

func TestTransformSchedJobGivesASequencesSteps(t *testing.T) {
	job := &entity.SchedJob{
		JobType: base.SchedJobTypeJobSequence,
		Sequence: &entity.SchedJobSequence{
			Mode: base.SchedJobSeqModeSequential, OnFailure: base.SchedJobSeqOnFailureContinue,
			Steps: []*entity.SchedJobSequenceStep{
				{Job: entity.ObjectID{ID: jobA}, Name: "migrate"},
				{Job: entity.ObjectID{ID: jobB}},
			},
		},
	}
	setting := &entity.Setting{ID: "seq", Type: base.SettingTypeSchedJob, Name: "release"}
	setting.MustSetData(job)
	refObjects := entity.NewRefObjects()
	refObjects.RefSettings[jobA] = &entity.Setting{ID: jobA, Type: base.SettingTypeSchedJob, Name: "db migrate",
		Status: base.SettingStatusActive, Scope: base.ObjectScopeApp, ObjectID: "app-1"}
	refObjects.RefApps["app-1"] = &entity.App{ID: "app-1", Name: "backend"}

	resp, err := TransformSchedJob(setting, refObjects, false)

	assert.NoError(t, err)
	assert.Nil(t, resp.Schedule, "no schedule")
	assert.Empty(t, resp.NextRuns)
	if assert.NotNil(t, resp.Sequence) && assert.Len(t, resp.Sequence.Steps, 2) {
		assert.Equal(t, base.SchedJobSeqOnFailureContinue, resp.Sequence.OnFailure)
		assert.Equal(t, "migrate", resp.Sequence.Steps[0].Name)
		assert.Equal(t, "db migrate", resp.Sequence.Steps[0].Job.Name)
		if assert.NotNil(t, resp.Sequence.Steps[0].App) {
			assert.Equal(t, "backend", resp.Sequence.Steps[0].App.Name, "whose job it is")
		}
		assert.Nil(t, resp.Sequence.Steps[1].App)
		assert.Equal(t, base.SettingStatusMissing, resp.Sequence.Steps[1].Job.Status, "a job not loaded")
		assert.Equal(t, jobB, resp.Sequence.Steps[1].Job.ID)
	}
}
