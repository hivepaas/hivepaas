package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0, base.AllDeploymentStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListDeploymentResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*DeploymentResp `json:"data"`
}

type DeploymentTransformInput struct {
	DeploymentInfoMap map[string]*cacheentity.DeploymentInfo
	TriggerUserMap    map[string]*entity.User
}

func TransformDeployments(
	deployments []*entity.Deployment,
	input *DeploymentTransformInput,
) (resp []*DeploymentResp, err error) {
	resp = make([]*DeploymentResp, 0, len(deployments))
	for _, deployment := range deployments {
		item, err := TransformDeployment(deployment, input)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
