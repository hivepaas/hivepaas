package registryauthrenewaldto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExecuteRegistryAuthRenewalReq struct {
	settings.GetUniqueSettingReq
	// TargetAuths are the credentials to renew; every Amazon ECR one when empty.
	TargetAuths basedto.ObjectIDSliceReq `json:"targetAuths"`
}

func NewExecuteRegistryAuthRenewalReq() *ExecuteRegistryAuthRenewalReq {
	return &ExecuteRegistryAuthRenewalReq{}
}

func (req *ExecuteRegistryAuthRenewalReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExecuteRegistryAuthRenewalResp struct {
	Meta *basedto.Meta                       `json:"meta"`
	Data *ExecuteRegistryAuthRenewalDataResp `json:"data"`
}

type ExecuteRegistryAuthRenewalDataResp struct {
	Task *basedto.ObjectIDResp `json:"task"`
}
