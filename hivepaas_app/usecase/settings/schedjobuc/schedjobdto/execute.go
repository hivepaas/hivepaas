package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSchedJobReq struct {
	settings.GetSettingReq
}

func NewExecuteSchedJobReq() *ExecuteSchedJobReq {
	return &ExecuteSchedJobReq{}
}

func (req *ExecuteSchedJobReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSchedJobResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *ExecuteSchedJobDataResp `json:"data"`
}

type ExecuteSchedJobDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
