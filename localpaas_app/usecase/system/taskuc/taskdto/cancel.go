package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type CancelTaskReq struct {
	ID string `json:"-"`
}

func NewCancelTaskReq() *CancelTaskReq {
	return &CancelTaskReq{}
}

func (req *CancelTaskReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CancelTaskResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *CancelTaskDataResp `json:"data"`
}

type CancelTaskDataResp struct {
	Canceled bool `json:"canceled"`
}
