package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RequestResetPasswordReq struct {
	ID                 string `json:"-"`
	SendResettingEmail bool   `json:"sendResettingEmail"`
}

func NewRequestResetPasswordReq() *RequestResetPasswordReq {
	return &RequestResetPasswordReq{}
}

func (req *RequestResetPasswordReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RequestResetPasswordResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *RequestResetPasswordDataResp `json:"data"`
}

type RequestResetPasswordDataResp struct {
	ResetPasswordLink string `json:"resetPasswordLink"`
}
