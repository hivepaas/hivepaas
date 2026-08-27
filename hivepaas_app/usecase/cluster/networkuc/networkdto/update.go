package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateNetworkReq struct {
	settings.UpdateSettingReq
}

func NewUpdateNetworkReq() *UpdateNetworkReq {
	return &UpdateNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNetworkReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
