package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UpdatePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func NewUpdatePasswordReq() *UpdatePasswordReq {
	return &UpdatePasswordReq{}
}

func (req *UpdatePasswordReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.CurrentPassword, true,
		1, userservice.PasswordRequirements.MaxLen, "currentPassword")...)
	validators = append(validators, basedto.ValidateStr(&req.NewPassword, true,
		userservice.PasswordRequirements.MinLen, userservice.PasswordRequirements.MaxLen, "newPassword")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdatePasswordResp struct {
	Meta *basedto.Meta `json:"meta"`
}
