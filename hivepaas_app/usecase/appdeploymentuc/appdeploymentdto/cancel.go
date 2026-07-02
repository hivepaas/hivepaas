package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CancelDeploymentResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *CancelDeploymentDataResp `json:"data"`
}

type CancelDeploymentDataResp struct {
	Canceled bool `json:"canceled"`
}
