package apikeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateAPIKeyStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateAPIKeyStatusReq() *UpdateAPIKeyStatusReq {
	return &UpdateAPIKeyStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAPIKeyStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAPIKeyStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
