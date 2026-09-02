package registryauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateRegistryAuthReq struct {
	settings.UpdateSettingReq
	*RegistryAuthBaseReq
}

func NewUpdateRegistryAuthReq() *UpdateRegistryAuthReq {
	return &UpdateRegistryAuthReq{}
}

func (req *UpdateRegistryAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateRegistryAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRegistryAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
