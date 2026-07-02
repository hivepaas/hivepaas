package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSchedJobReq struct {
	settings.GetSettingReq
}

func NewExecuteSchedJobReq() *ExecuteSchedJobReq {
	return &ExecuteSchedJobReq{}
}

func (req *ExecuteSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSchedJobResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *ExecuteSchedJobDataResp `json:"data"`
}

type ExecuteSchedJobDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
