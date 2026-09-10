package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListTargetObjectsReq struct {
	Scope  *entity.ObjectScope `json:"-" mapstructure:"-"`
	Paging basedto.Paging      `json:"-"`
}

func NewListTargetObjectsReq() *ListTargetObjectsReq {
	return &ListTargetObjectsReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListTargetObjectsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListTargetObjectsResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*TargetObjectResp `json:"data"`
}

type TargetObjectResp struct {
	ID   string              `json:"id"`
	Type base.TaskTargetType `json:"type"`
	Name string              `json:"name"`
}
