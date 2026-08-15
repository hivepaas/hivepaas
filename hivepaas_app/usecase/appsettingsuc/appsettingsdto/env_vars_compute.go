package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ComputeAppEnvVarsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	*AppEnvVarsBaseReq
}

func NewComputeAppEnvVarsReq() *ComputeAppEnvVarsReq {
	return &ComputeAppEnvVarsReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *ComputeAppEnvVarsReq) ModifyRequest() error {
	return req.modifyRequest()
}

func (req *ComputeAppEnvVarsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ComputeAppEnvVarsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data []*basedto.EnvVarResp `json:"data"`
}
