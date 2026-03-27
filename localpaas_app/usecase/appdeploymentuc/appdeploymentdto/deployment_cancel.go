package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type CancelDeploymentReq struct {
	ProjectID    string `json:"-"`
	AppID        string `json:"-"`
	DeploymentID string `json:"-"`
}

func NewCancelDeploymentReq() *CancelDeploymentReq {
	return &CancelDeploymentReq{}
}

func (req *CancelDeploymentReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CancelDeploymentResp struct {
	Meta *basedto.Meta `json:"meta"`
}
