package imagescandto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteImageScanReq struct {
	settings.GetUniqueSettingReq
}

func NewExecuteImageScanReq() *ExecuteImageScanReq {
	return &ExecuteImageScanReq{}
}

func (req *ExecuteImageScanReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteImageScanResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *ExecuteImageScanDataResp `json:"data"`
}

type ExecuteImageScanDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
