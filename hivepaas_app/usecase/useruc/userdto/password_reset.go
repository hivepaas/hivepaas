package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ResetPasswordReq struct {
	ID       string `json:"-"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func NewResetPasswordReq() *ResetPasswordReq {
	return &ResetPasswordReq{}
}

func (req *ResetPasswordReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateStr(&req.Password, true, nameMinLen, nameMaxLen, "password")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ResetPasswordResp struct {
	Meta *basedto.Meta `json:"meta"`
}
