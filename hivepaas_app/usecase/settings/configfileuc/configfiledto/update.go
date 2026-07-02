package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateConfigFileReq struct {
	settings.UpdateSettingReq
	*ConfigFileBaseReq
}

func NewUpdateConfigFileReq() *UpdateConfigFileReq {
	return &UpdateConfigFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateConfigFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate(false, "")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateConfigFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
