package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateConfigFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate(false, "")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateConfigFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
