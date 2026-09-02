package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSecretStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateSecretStatusReq() *UpdateSecretStatusReq {
	return &UpdateSecretStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSecretStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingStatusReq.Validate()...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSecretStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
