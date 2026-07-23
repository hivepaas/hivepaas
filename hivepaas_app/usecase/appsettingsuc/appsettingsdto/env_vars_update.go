package appsettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppEnvVarsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	UpdateVer int    `json:"updateVer"`
	*AppEnvVarsBaseReq
}

type AppEnvVarsBaseReq struct {
	BuildtimeEnvVars []*basedto.EnvVarReq `json:"buildtimeEnvVars"`
	RuntimeEnvVars   []*basedto.EnvVarReq `json:"runtimeEnvVars"`
	SharedEnvVars    []*basedto.EnvVarReq `json:"sharedEnvVars"`
}

func (req *AppEnvVarsBaseReq) modifyRequest() error {
	for _, env := range req.BuildtimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	for _, env := range req.RuntimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	for _, env := range req.SharedEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	return nil
}

func (req *AppEnvVarsBaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}

	res = append(res, basedto.ValidateEnvVarsReq(req.BuildtimeEnvVars, field+"buildtimeEnvVars")...)
	res = append(res, basedto.ValidateEnvVarsReq(req.RuntimeEnvVars, field+"runtimeEnvVars")...)
	res = append(res, basedto.ValidateEnvVarsReq(req.SharedEnvVars, field+"sharedEnvVars")...)

	allSharedEnvs := make(map[string]struct{}, len(req.SharedEnvVars))
	for _, env := range req.SharedEnvVars {
		allSharedEnvs[env.Key] = struct{}{}
	}
	for _, env := range req.RuntimeEnvVars {
		if _, ok := allSharedEnvs[env.Key]; ok {
			res = append(res, vld.Must(false).OnError(
				vld.SetField("runtimeEnvVars", nil),
				vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
			))
			continue
		}
		if !base.IsAppRuntimeEnvAllowed(env.Key) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField("runtimeEnvVars", nil),
				vld.SetCustomKey("ERR_VLD_VALUE_RESERVED"),
				vld.SetParam("Value", env.Key),
			))
		}
	}

	return res
}

func NewUpdateAppEnvVarsReq() *UpdateAppEnvVarsReq {
	return &UpdateAppEnvVarsReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *UpdateAppEnvVarsReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
