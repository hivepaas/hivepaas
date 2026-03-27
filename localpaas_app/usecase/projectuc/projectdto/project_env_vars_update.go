package projectdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateProjectEnvVarsReq struct {
	ProjectID        string               `json:"-"`
	BuildtimeEnvVars []*basedto.EnvVarReq `json:"buildtimeEnvVars"`
	RuntimeEnvVars   []*basedto.EnvVarReq `json:"runtimeEnvVars"`
	UpdateVer        int                  `json:"updateVer"`
}

func NewUpdateProjectEnvVarsReq() *UpdateProjectEnvVarsReq {
	return &UpdateProjectEnvVarsReq{}
}

func (req *UpdateProjectEnvVarsReq) ModifyRequest() error {
	for _, env := range req.BuildtimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	for _, env := range req.RuntimeEnvVars {
		env.Key = strings.TrimSpace(env.Key)
		env.Value = strings.TrimSpace(env.Value)
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateEnvVarsReq(req.BuildtimeEnvVars, "buildtimeEnvVars")...)
	validators = append(validators, basedto.ValidateEnvVarsReq(req.RuntimeEnvVars, "runtimeEnvVars")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectEnvVarsResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *UpdateProjectEnvVarsDataResp `json:"data"`
}

type UpdateProjectEnvVarsDataResp struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
