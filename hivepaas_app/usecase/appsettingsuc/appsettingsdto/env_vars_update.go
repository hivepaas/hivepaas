package appsettingsdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppEnvVarsReq struct {
	ProjectID        string               `json:"-"`
	AppID            string               `json:"-"`
	BuildtimeEnvVars []*basedto.EnvVarReq `json:"buildtimeEnvVars"`
	RuntimeEnvVars   []*basedto.EnvVarReq `json:"runtimeEnvVars"`
	SharedEnvVars    []*basedto.EnvVarReq `json:"sharedEnvVars"`
	UpdateVer        int                  `json:"updateVer"`
}

func NewUpdateAppEnvVarsReq() *UpdateAppEnvVarsReq {
	return &UpdateAppEnvVarsReq{}
}

func (req *UpdateAppEnvVarsReq) ModifyRequest() error {
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

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateEnvVarsReq(req.BuildtimeEnvVars, "buildtimeEnvVars")...)
	validators = append(validators, basedto.ValidateEnvVarsReq(req.RuntimeEnvVars, "runtimeEnvVars")...)
	validators = append(validators, basedto.ValidateEnvVarsReq(req.SharedEnvVars, "sharedEnvVars")...)

	allSharedEnvs := make(map[string]struct{}, len(req.SharedEnvVars))
	for _, env := range req.SharedEnvVars {
		allSharedEnvs[env.Key] = struct{}{}
	}
	for _, env := range req.RuntimeEnvVars {
		if _, ok := allSharedEnvs[env.Key]; ok {
			validators = append(validators, vld.Must(false).OnError(
				vld.SetField("runtimeEnvVars", nil),
				vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
			))
			continue
		}
		if !base.IsAppRuntimeEnvAllowed(env.Key) {
			validators = append(validators, vld.Must(false).OnError(
				vld.SetField("runtimeEnvVars", nil),
				vld.SetCustomKey("ERR_VLD_VALUE_RESERVED"),
				vld.SetParam("Value", env.Key),
			))
		}
	}

	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
