package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
)

type LoginPasswordForgotReq struct {
	Email string `json:"email"`
}

func NewLoginPasswordForgotReq() *LoginPasswordForgotReq {
	return &LoginPasswordForgotReq{}
}

func (req *LoginPasswordForgotReq) ModifyRequest() error {
	req.Email = strutil.NormalizeEmail(req.Email)
	return nil
}

func (req *LoginPasswordForgotReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Email, true, 1,
		maxEmailLen, "email")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginPasswordForgotResp struct {
	Meta *basedto.Meta `json:"meta"`
}
