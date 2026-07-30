package projectenvsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ComputeProjectEnvEnvVarsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	*ProjectEnvVarsBaseReq
}

func NewComputeProjectEnvEnvVarsReq() *ComputeProjectEnvEnvVarsReq {
	return &ComputeProjectEnvEnvVarsReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *ComputeProjectEnvEnvVarsReq) ModifyRequest() error {
	return req.modifyRequest()
}

func (req *ComputeProjectEnvEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ComputeProjectEnvEnvVarsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data []*basedto.EnvVarResp `json:"data"`
}
