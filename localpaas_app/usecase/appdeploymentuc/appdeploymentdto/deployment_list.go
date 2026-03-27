package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
)

type ListDeploymentReq struct {
	ProjectID string                  `json:"-"`
	AppID     string                  `json:"-"`
	Status    []base.DeploymentStatus `json:"-" mapstructure:"status"`
	Search    string                  `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListDeploymentReq() *ListDeploymentReq {
	return &ListDeploymentReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListDeploymentReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0, base.AllDeploymentStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListDeploymentResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*DeploymentResp `json:"data"`
}

func TransformDeployments(deployments []*entity.Deployment, deploymentInfoMap map[string]*cacheentity.DeploymentInfo) (
	resp []*DeploymentResp, err error) {
	resp = make([]*DeploymentResp, 0, len(deployments))
	for _, task := range deployments {
		item, err := TransformDeployment(task, deploymentInfoMap[task.ID])
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
