package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAcmeDnsProviderReq struct {
	settings.UpdateSettingReq
	*AcmeDnsProviderBaseReq
}

func NewUpdateAcmeDnsProviderReq() *UpdateAcmeDnsProviderReq {
	return &UpdateAcmeDnsProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAcmeDnsProviderReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAcmeDnsProviderResp struct {
	Meta *basedto.Meta `json:"meta"`
}
