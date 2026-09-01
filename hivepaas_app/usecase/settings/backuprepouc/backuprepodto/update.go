package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateBackupRepoReq struct {
	settings.UpdateSettingReq
	*BackupRepoBaseReq
}

func NewUpdateBackupRepoReq() *UpdateBackupRepoReq {
	return &UpdateBackupRepoReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateBackupRepoResp struct {
	Meta *basedto.Meta `json:"meta"`
}
