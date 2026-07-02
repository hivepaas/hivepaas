package registryauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
func (req *UpdateRegistryAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRegistryAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
