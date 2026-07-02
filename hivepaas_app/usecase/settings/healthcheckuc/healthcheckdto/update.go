package healthcheckdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateHealthcheckReq struct {
	settings.UpdateSettingReq
	*HealthcheckBaseReq
}

func NewUpdateHealthcheckReq() *UpdateHealthcheckReq {
	return &UpdateHealthcheckReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateHealthcheckReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateHealthcheckResp struct {
	Meta *basedto.Meta `json:"meta"`
}
