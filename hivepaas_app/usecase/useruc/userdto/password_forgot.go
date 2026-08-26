package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type PasswordForgotReq struct {
	Email string `json:"email"`
}

func NewPasswordForgotReq() *PasswordForgotReq {
	return &PasswordForgotReq{}
}

func (req *PasswordForgotReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateEmail(&req.Email, true, "email")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type PasswordForgotResp struct {
	Meta *basedto.Meta `json:"meta"`
}
