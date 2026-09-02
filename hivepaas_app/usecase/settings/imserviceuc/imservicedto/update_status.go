package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateIMServiceStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateIMServiceStatusReq() *UpdateIMServiceStatusReq {
	return &UpdateIMServiceStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateIMServiceStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingStatusReq.Validate()...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateIMServiceStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
