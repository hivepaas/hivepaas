package taskdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type CancelTaskReq struct {
	ID string `json:"-"`
}

func NewCancelTaskReq() *CancelTaskReq {
	return &CancelTaskReq{}
}

func (req *CancelTaskReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
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
