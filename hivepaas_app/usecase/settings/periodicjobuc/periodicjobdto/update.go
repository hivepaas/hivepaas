package periodicjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdatePeriodicJobReq struct {
	settings.UpdateSettingReq
	*PeriodicJobBaseReq
}

func NewUpdatePeriodicJobReq() *UpdatePeriodicJobReq {
	return &UpdatePeriodicJobReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdatePeriodicJobReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdatePeriodicJobResp struct {
	Meta *basedto.Meta `json:"meta"`
}
