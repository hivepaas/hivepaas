package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/backup"
)

type UpdateBackupRepoReq struct {
	settings.UpdateSettingReq
	*BackupRepoBaseUpdateReq
}

func NewUpdateBackupRepoReq() *UpdateBackupRepoReq {
	return &UpdateBackupRepoReq{}
}

// Validate (NOTE: this doesn't implement interface basedto.ReqValidator)
func (req *UpdateBackupRepoReq) Validate(engine backup.EngineType) hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("", engine)...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateBackupRepoResp struct {
	Meta *basedto.Meta `json:"meta"`
}
