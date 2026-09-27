package schedjobdto

import (
	"fmt"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const sequenceStepNameMaxLen = 100

// SchedJobSequenceReq is a job sequence's jobs and how it runs them.
type SchedJobSequenceReq struct {
	Mode      base.SchedJobSeqMode       `json:"mode"`
	OnFailure base.SchedJobSeqOnFailure  `json:"onFailure"`
	Steps     []*SchedJobSequenceStepReq `json:"steps"`
}

type SchedJobSequenceStepReq struct {
	Job  basedto.ObjectIDReq `json:"job"`
	Name string              `json:"name"`
}

func (req *SchedJobSequenceReq) ToEntity() *entity.SchedJobSequence {
	if req == nil {
		return nil
	}
	seq := &entity.SchedJobSequence{Mode: req.Mode, OnFailure: req.OnFailure}
	for _, step := range req.Steps {
		if step == nil {
			continue
		}
		seq.Steps = append(seq.Steps, &entity.SchedJobSequenceStep{
			Job:  entity.ObjectID{ID: step.Job.ID},
			Name: step.Name,
		})
	}
	return seq
}

func (req *SchedJobSequenceReq) modifyRequest() {
	if req == nil {
		return
	}
	if req.Mode == "" {
		req.Mode = base.SchedJobSeqModeSequential
	}
	if req.OnFailure == "" {
		req.OnFailure = base.SchedJobSeqOnFailureStop
	}
	for _, step := range req.Steps {
		if step != nil {
			step.Name = strings.TrimSpace(step.Name)
		}
	}
}

func (req *SchedJobSequenceReq) validate(field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateCond(req != nil, field)...)
	if req == nil {
		return res
	}
	field += "."
	res = append(res, basedto.ValidateStrIn(&req.Mode, true, base.AllSchedJobSeqModes, field+"mode")...)
	res = append(res, basedto.ValidateStrIn(&req.OnFailure, true, base.AllSchedJobSeqOnFailures,
		field+"onFailure")...)
	res = append(res, basedto.ValidateCond(len(req.Steps) >= 1 && len(req.Steps) <= base.SchedJobSeqMaxSteps,
		field+"steps")...)
	for i, step := range req.Steps {
		stepField := fmt.Sprintf("%ssteps[%d]", field, i)
		res = append(res, basedto.ValidateCond(step != nil, stepField)...)
		if step == nil {
			continue
		}
		res = append(res, basedto.ValidateObjectIDReq(&step.Job, true, stepField+".job")...)
		res = append(res, basedto.ValidateStr(&step.Name, false, 0, sequenceStepNameMaxLen, stepField+".name")...)
	}
	return res
}

// validateSequenceFields is what a job's type says about its sequence: a
// job-sequence has one and runs nothing of its own, any other type has none.
func (req *SchedJobBaseReq) validateSequenceFields(field string) (res []vld.Validator) {
	if req.JobType != base.SchedJobTypeJobSequence {
		return basedto.ValidateCond(req.Sequence == nil, field+"sequence")
	}
	res = append(res, req.Sequence.validate(field+"sequence")...)
	res = append(res, basedto.ValidateCond(req.Command == nil, field+"command")...)
	res = append(res, basedto.ValidateCond(req.App.ID == "", field+"app")...)
	res = append(res, basedto.ValidateCond(req.CommandOutput == nil, field+"commandOutput")...)
	return res
}

type SchedJobSequenceResp struct {
	Mode      base.SchedJobSeqMode        `json:"mode"`
	OnFailure base.SchedJobSeqOnFailure   `json:"onFailure"`
	Steps     []*SchedJobSequenceStepResp `json:"steps"`
}

type SchedJobSequenceStepResp struct {
	// Job is the scheduled job the step runs; "missing" when it is gone.
	Job *settings.BaseSettingResp `json:"job"`
	// App is the app the job belongs to, when it is an app's.
	App  *basedto.NamedObjectResp `json:"app,omitempty"`
	Name string                   `json:"name,omitempty"`
}

// TransformSchedJobSequence is a sequence with each step's job named, from the
// jobs refObjects holds.
func TransformSchedJobSequence(seq *entity.SchedJobSequence, refObjects *entity.RefObjects) *SchedJobSequenceResp {
	if seq == nil {
		return nil
	}
	resp := &SchedJobSequenceResp{
		Mode:      seq.Mode,
		OnFailure: seq.OnFailure,
		Steps:     make([]*SchedJobSequenceStepResp, 0, len(seq.Steps)),
	}
	for _, step := range seq.Steps {
		var member *entity.Setting
		if refObjects != nil {
			member = refObjects.RefSettings[step.Job.ID]
		}
		jobResp, _ := settings.TransformSettingBase(member)
		if jobResp == nil {
			jobResp = settings.NewMissingSetting(step.Job.ID, base.SettingTypeSchedJob)
		}
		stepResp := &SchedJobSequenceStepResp{Job: jobResp, Name: step.Name}
		if member != nil && member.Scope == base.ObjectScopeApp {
			if app := refObjects.RefApps[member.ObjectID]; app != nil {
				stepResp.App = &basedto.NamedObjectResp{ID: app.ID, Name: app.Name}
			}
		}
		resp.Steps = append(resp.Steps, stepResp)
	}
	return resp
}
