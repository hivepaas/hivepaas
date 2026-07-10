package sslrenewaldto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteSSLRenewalReq struct {
	settings.GetUniqueSettingReq
	TargetSSLs basedto.ObjectIDSliceReq `json:"targetSSLs"`
}

func NewExecuteSSLRenewalReq() *ExecuteSSLRenewalReq {
	return &ExecuteSSLRenewalReq{}
}

func (req *ExecuteSSLRenewalReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteSSLRenewalResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *ExecuteSSLRenewalDataResp `json:"data"`
}

type ExecuteSSLRenewalDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
