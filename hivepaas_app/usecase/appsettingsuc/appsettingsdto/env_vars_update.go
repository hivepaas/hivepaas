package appsettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateAppEnvVarsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	UpdateVer    int    `json:"updateVer"`
	*AppEnvVarsBaseReq
}

type AppEnvVarsBaseReq struct {
	RuntimeEnvVars   []*basedto.EnvVarReq `json:"runtimeEnvVars"`
	SharedEnvVars    []*basedto.EnvVarReq `json:"sharedEnvVars"`
	BuildtimeEnvVars []*basedto.EnvVarReq `json:"buildtimeEnvVars"`
}

func (req *AppEnvVarsBaseReq) modifyRequest() error {
	for _, env := range req.RuntimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	for _, env := range req.BuildtimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}

	adjSharedEnvVars := make([]*basedto.EnvVarReq, 0, len(req.SharedEnvVars))
	for _, env := range req.SharedEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
		if base.IsAppSharedEnvSettable(env.Key) {
			adjSharedEnvVars = append(adjSharedEnvVars, env)
		}
	}
	req.SharedEnvVars = adjSharedEnvVars
	return nil
}

func (req *AppEnvVarsBaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}

	res = append(res, basedto.ValidateEnvVarsReq(req.RuntimeEnvVars, field+"runtimeEnvVars")...)
	res = append(res, basedto.ValidateEnvVarsReq(req.SharedEnvVars, field+"sharedEnvVars")...)
	res = append(res, basedto.ValidateEnvVarsReq(req.BuildtimeEnvVars, field+"buildtimeEnvVars")...)

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
	for _, env := range req.SharedEnvVars {
		if !base.IsAppSharedEnvAllowed(env.Key) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField("sharedEnvVars", nil),
				vld.SetCustomKey("ERR_VLD_VALUE_RESERVED"),
				vld.SetParam("Value", env.Key),
			))
		}
	}
	for _, env := range req.BuildtimeEnvVars {
		if !base.IsAppBuildEnvAllowed(env.Key) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField("buildtimeEnvVars", nil),
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
func (req *UpdateAppEnvVarsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
