package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateNetworkStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateNetworkStatusReq() *UpdateNetworkStatusReq {
	return &UpdateNetworkStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNetworkStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNetworkStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
