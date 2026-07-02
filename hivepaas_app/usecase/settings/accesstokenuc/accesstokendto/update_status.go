package accesstokendto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAccessTokenStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateAccessTokenStatusReq() *UpdateAccessTokenStatusReq {
	return &UpdateAccessTokenStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAccessTokenStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAccessTokenStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
