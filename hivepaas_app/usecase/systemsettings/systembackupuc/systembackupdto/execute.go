package systembackupdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSystemBackupReq struct {
	settings.GetUniqueSettingReq
}

func NewExecuteSystemBackupReq() *ExecuteSystemBackupReq {
	return &ExecuteSystemBackupReq{}
}

func (req *ExecuteSystemBackupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSystemBackupResp struct {
	Meta *basedto.Meta                `json:"meta"`
	Data *ExecuteSystemBackupDataResp `json:"data"`
}

type ExecuteSystemBackupDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
