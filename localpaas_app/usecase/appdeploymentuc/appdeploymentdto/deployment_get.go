package appdeploymentdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type GetDeploymentReq struct {
	ProjectID    string `json:"-"`
	AppID        string `json:"-"`
	DeploymentID string `json:"-"`
}

func NewGetDeploymentReq() *GetDeploymentReq {
	return &GetDeploymentReq{}
}

func (req *GetDeploymentReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *DeploymentResp `json:"data"`
}

type DeploymentResp struct {
	ID        string                        `json:"id"`
	Status    base.DeploymentStatus         `json:"status"`
	UpdateVer int                           `json:"updateVer"`
	Settings  *entity.AppDeploymentSettings `json:"settings"`
	Output    *entity.AppDeploymentOutput   `json:"output"`

	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TransformDeployment(deployment *entity.Deployment, deploymentInfo *cacheentity.DeploymentInfo) (
	resp *DeploymentResp, err error) {
	if err = copier.Copy(&resp, &deployment); err != nil {
		return nil, apperrors.Wrap(err)
	}
	if deploymentInfo != nil {
		resp.Status = deploymentInfo.Status
		if deploymentInfo.Status == base.DeploymentStatusInProgress {
			resp.StartedAt = deploymentInfo.StartedAt
		}
	}
	return resp, nil
}
