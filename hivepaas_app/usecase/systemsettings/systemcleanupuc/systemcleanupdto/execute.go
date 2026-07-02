package systemcleanupdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSystemCleanupReq struct {
	settings.GetSettingReq
}

func NewExecuteSystemCleanupReq() *ExecuteSystemCleanupReq {
	return &ExecuteSystemCleanupReq{}
}

func (req *ExecuteSystemCleanupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSystemCleanupResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *ExecuteSystemCleanupDataResp `json:"data"`
}

type ExecuteSystemCleanupDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
