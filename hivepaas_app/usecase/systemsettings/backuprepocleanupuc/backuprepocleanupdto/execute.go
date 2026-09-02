package backuprepocleanupdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteBackupRepoCleanupReq struct {
	settings.GetUniqueSettingReq
	TargetRepos basedto.ObjectIDSliceReq `json:"targetRepos"`
}

func NewExecuteBackupRepoCleanupReq() *ExecuteBackupRepoCleanupReq {
	return &ExecuteBackupRepoCleanupReq{}
}

func (req *ExecuteBackupRepoCleanupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	// An empty list means every repository, so the IDs are optional - but the ones given must be
	// well formed and not repeated, otherwise the run silently targets fewer repos than asked.
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.TargetRepos, true, 0,
		"targetRepos")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteBackupRepoCleanupResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *ExecuteBackupRepoCleanupDataResp `json:"data"`
}

type ExecuteBackupRepoCleanupDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
