package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateNetworkReq struct {
	settings.UpdateSettingReq
}

func NewUpdateNetworkReq() *UpdateNetworkReq {
	return &UpdateNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
