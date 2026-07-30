package projectsettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateProjectEnvVarsReq struct {
	ProjectID string `json:"-"`
	UpdateVer int    `json:"updateVer"`
	*ProjectEnvVarsBaseReq
}

type ProjectEnvVarsBaseReq struct {
	RuntimeEnvVars   []*basedto.EnvVarReq `json:"runtimeEnvVars"`
	BuildtimeEnvVars []*basedto.EnvVarReq `json:"buildtimeEnvVars"`
}

func (req *ProjectEnvVarsBaseReq) modifyRequest() error {
	for _, env := range req.RuntimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	for _, env := range req.BuildtimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *ProjectEnvVarsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateEnvVarsReq(req.RuntimeEnvVars, field+"runtimeEnvVars")...)
	res = append(res, basedto.ValidateEnvVarsReq(req.BuildtimeEnvVars, field+"buildtimeEnvVars")...)

	for _, env := range req.RuntimeEnvVars {
		if !base.IsAppRuntimeEnvAllowed(env.Key) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField("runtimeEnvVars", nil),
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

func NewUpdateProjectEnvVarsReq() *UpdateProjectEnvVarsReq {
	return &UpdateProjectEnvVarsReq{}
}

func (req *UpdateProjectEnvVarsReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
