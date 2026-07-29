package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ComputeProjectEnvVarsReq struct {
	ProjectID string `json:"-"`
	*ProjectEnvVarsBaseReq
}

func NewComputeProjectEnvVarsReq() *ComputeProjectEnvVarsReq {
	return &ComputeProjectEnvVarsReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *ComputeProjectEnvVarsReq) ModifyRequest() error {
	return req.modifyRequest()
}

func (req *ComputeProjectEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ComputeProjectEnvVarsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data []*basedto.EnvVarResp `json:"data"`
}
