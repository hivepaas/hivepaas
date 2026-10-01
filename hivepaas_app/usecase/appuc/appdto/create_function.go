package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// CreateFunctionReq creates a function: an app of kind function, built from its
// source and deployed as soon as it exists.
type CreateFunctionReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	*AppBaseReq
	// Source is what the function's deployment settings take: its code, its
	// runtime and its limits.
	Source *appsettingsdto.DeploymentFunctionSourceReq `json:"source"`
}

func NewCreateFunctionReq() *CreateFunctionReq {
	return &CreateFunctionReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *CreateFunctionReq) ModifyRequest() error {
	// A body without the app's fields has them empty, which is refused below.
	if req.AppBaseReq == nil {
		req.AppBaseReq = &AppBaseReq{}
	}
	if err := req.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if req.Source != nil {
		req.Source.Normalize()
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateFunctionReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, req.validate("")...)
	validators = append(validators, vld.Must(req.Source != nil).OnError(
		vld.SetField("source", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
	))
	validators = append(validators, req.Source.Validate("source")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateFunctionResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *CreateFunctionDataResp `json:"data"`
}

// CreateFunctionDataResp is the function created, and its first deployment.
type CreateFunctionDataResp struct {
	ID           string `json:"id"`
	DeploymentID string `json:"deploymentId"`
	TaskID       string `json:"taskId"`
}
