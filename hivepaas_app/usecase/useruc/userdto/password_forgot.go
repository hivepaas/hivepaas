package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type PasswordForgotReq struct {
	Email string `json:"email"`
}

func NewPasswordForgotReq() *PasswordForgotReq {
	return &PasswordForgotReq{}
}

func (req *PasswordForgotReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateEmail(&req.Email, true, "email")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type PasswordForgotResp struct {
	Meta *basedto.Meta `json:"meta"`
}
