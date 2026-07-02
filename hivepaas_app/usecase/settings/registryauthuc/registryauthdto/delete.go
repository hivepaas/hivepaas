package registryauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteRegistryAuthReq struct {
	settings.DeleteSettingReq
}

func NewDeleteRegistryAuthReq() *DeleteRegistryAuthReq {
	return &DeleteRegistryAuthReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteRegistryAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteRegistryAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
