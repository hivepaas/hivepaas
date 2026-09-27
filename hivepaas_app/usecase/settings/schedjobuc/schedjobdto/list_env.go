package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ListEnvSchedJobReq lists a project env's scheduled jobs with those of the
// apps in it.
type ListEnvSchedJobReq struct {
	ListSchedJobReq
	JobTypes []base.SchedJobType `json:"-" mapstructure:"jobType"`
	AppID    string              `json:"-" mapstructure:"appId"`
}

func NewListEnvSchedJobReq() *ListEnvSchedJobReq {
	return &ListEnvSchedJobReq{ListSchedJobReq: *NewListSchedJobReq()}
}

func (req *ListEnvSchedJobReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateSlice(req.JobTypes, true, 0, base.AllSchedJobTypes,
		"jobType")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, false, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListEnvSchedJobResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*EnvSchedJobResp `json:"data"`
}

// EnvSchedJobResp is a scheduled job in an env's list: the env's own, or an
// app's, with the app it belongs to.
type EnvSchedJobResp struct {
	*SchedJobResp
	Scope    base.ObjectScopeType     `json:"scope"`
	OwnerApp *basedto.NamedObjectResp `json:"ownerApp,omitempty"`
}

func TransformEnvSchedJobs(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) ([]*EnvSchedJobResp, error) {
	resp := make([]*EnvSchedJobResp, 0, len(settings))
	for _, setting := range settings {
		job, err := TransformSchedJob(setting, refObjects, true)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		item := &EnvSchedJobResp{SchedJobResp: job, Scope: setting.Scope}
		if setting.Scope == base.ObjectScopeApp {
			if app := refObjects.RefApps[setting.ObjectID]; app != nil {
				item.OwnerApp = &basedto.NamedObjectResp{ID: app.ID, Name: app.Name}
			}
		}
		resp = append(resp, item)
	}
	return resp, nil
}
