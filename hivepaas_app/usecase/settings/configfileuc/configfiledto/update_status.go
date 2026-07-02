package configfiledto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateConfigFileStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateConfigFileStatusReq() *UpdateConfigFileStatusReq {
	return &UpdateConfigFileStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateConfigFileStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateConfigFileStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
