package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateTaskStatusReq struct {
	ID        string           `json:"-"`
	Status    *base.TaskStatus `json:"status"`
	UpdateVer int              `json:"updateVer"`
}

func NewUpdateTaskStatusReq() *UpdateTaskStatusReq {
	return &UpdateTaskStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateTaskStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllTaskSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateTaskStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
