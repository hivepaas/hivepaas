package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
)

type GetDeploymentStatusReq struct {
	ProjectID    string `json:"-"`
	AppID        string `json:"-"`
	DeploymentID string `json:"-"`
}

func NewGetDeploymentStatusReq() *GetDeploymentStatusReq {
	return &GetDeploymentStatusReq{}
}

func (req *GetDeploymentStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentStatusResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *DeploymentStatusResp `json:"data"`
}

type DeploymentStatusResp struct {
	Status base.DeploymentStatus `json:"status"`
}

func TransformDeploymentStatus(
	deployment *entity.Deployment,
	deploymentInfo *cacheentity.DeploymentInfo,
) *DeploymentStatusResp {
	resp := &DeploymentStatusResp{
		Status: deployment.Status,
	}
	if deploymentInfo != nil {
		resp.Status = deploymentInfo.Status
	}
	return resp
}
