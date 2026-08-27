package sessiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *LoginPasswordForgotReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Email, true, 1,
		maxEmailLen, "email")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type LoginPasswordForgotResp struct {
	Meta *basedto.Meta `json:"meta"`
}
