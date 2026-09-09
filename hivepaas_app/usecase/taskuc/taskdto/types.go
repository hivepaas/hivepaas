package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListTaskTypeReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewListTaskTypeReq() *ListTaskTypeReq {
	return &ListTaskTypeReq{}
}

func (req *ListTaskTypeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListTaskTypeResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []base.TaskType   `json:"data"`
}
