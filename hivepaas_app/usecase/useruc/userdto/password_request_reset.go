package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type RequestResetPasswordReq struct {
	ID                 string `json:"-"`
	SendResettingEmail bool   `json:"sendResettingEmail"`
}

func NewRequestResetPasswordReq() *RequestResetPasswordReq {
	return &RequestResetPasswordReq{}
}

func (req *RequestResetPasswordReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type RequestResetPasswordResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *RequestResetPasswordDataResp `json:"data"`
}

type RequestResetPasswordDataResp struct {
	ResetPasswordLink string `json:"resetPasswordLink"`
}
