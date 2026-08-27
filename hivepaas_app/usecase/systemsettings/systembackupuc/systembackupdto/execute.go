package systembackupdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSystemBackupReq struct {
	settings.GetUniqueSettingReq
}

func NewExecuteSystemBackupReq() *ExecuteSystemBackupReq {
	return &ExecuteSystemBackupReq{}
}

func (req *ExecuteSystemBackupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSystemBackupResp struct {
	Meta *basedto.Meta                `json:"meta"`
	Data *ExecuteSystemBackupDataResp `json:"data"`
}

type ExecuteSystemBackupDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
