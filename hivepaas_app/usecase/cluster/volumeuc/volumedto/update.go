package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateVolumeReq struct {
	settings.UpdateSettingReq
}

func NewUpdateVolumeReq() *UpdateVolumeReq {
	return &UpdateVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateVolumeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
