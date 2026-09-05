package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ChangeRepoPasswordReq struct {
	settings.UpdateSettingReq
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func NewChangeRepoPasswordReq() *ChangeRepoPasswordReq {
	return &ChangeRepoPasswordReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ChangeRepoPasswordReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateStr(&req.CurrentPassword, true,
		1, backupreposervice.PasswordRequirements.MaxLen, "currentPassword")...)
	validators = append(validators, basedto.ValidateStr(&req.NewPassword, true,
		backupreposervice.PasswordRequirements.MinLen, backupreposervice.PasswordRequirements.MaxLen,
		"newPassword")...)
	validators = append(validators, basedto.ValidateCond(req.NewPassword != req.CurrentPassword,
		"newPassword")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ChangeRepoPasswordResp struct {
	Meta *basedto.Meta `json:"meta"`
}
