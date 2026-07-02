package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	passwordMinLen = 8
	passwordMaxLen = 64
)

type UpdatePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func NewUpdatePasswordReq() *UpdatePasswordReq {
	return &UpdatePasswordReq{}
}

func (req *UpdatePasswordReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.CurrentPassword, true,
		1, passwordMaxLen, "currentPassword")...)
	validators = append(validators, basedto.ValidateStr(&req.NewPassword, true,
		passwordMinLen, passwordMaxLen, "newPassword")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdatePasswordResp struct {
	Meta *basedto.Meta `json:"meta"`
}
