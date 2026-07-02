package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAcmeDnsProviderStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateAcmeDnsProviderStatusReq() *UpdateAcmeDnsProviderStatusReq {
	return &UpdateAcmeDnsProviderStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAcmeDnsProviderStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAcmeDnsProviderStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
