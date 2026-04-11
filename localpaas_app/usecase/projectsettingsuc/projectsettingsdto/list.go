package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type ListProjectReq struct {
	Status []base.ProjectStatus `json:"-" mapstructure:"status"`
	Search string               `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListProjectReq() *ListProjectReq {
	return &ListProjectReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllProjectStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListProjectResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*ProjectResp    `json:"data"`
}

func TransformProjects(projects []*entity.Project) ([]*ProjectResp, error) {
	return basedto.TransformObjectSlice(projects, TransformProject) //nolint:wrapcheck
}
