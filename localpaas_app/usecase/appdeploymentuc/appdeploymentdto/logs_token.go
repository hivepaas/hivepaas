package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetDeploymentLogsTokenReq struct {
	ProjectID    string `json:"-"`
	AppID        string `json:"-"`
	DeploymentID string `json:"-"`
}

func NewGetDeploymentLogsTokenReq() *GetDeploymentLogsTokenReq {
	return &GetDeploymentLogsTokenReq{}
}

func (req *GetDeploymentLogsTokenReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentLogsTokenResp struct {
	Meta *basedto.Meta                `json:"meta"`
	Data *DeploymentLogsTokenDataResp `json:"data"`
}

type DeploymentLogsTokenDataResp struct {
	Token string `json:"token"`
}
