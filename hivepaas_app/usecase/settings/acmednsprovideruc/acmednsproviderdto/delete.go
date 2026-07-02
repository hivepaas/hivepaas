package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAcmeDnsProviderReq struct {
	settings.DeleteSettingReq
}

func NewDeleteAcmeDnsProviderReq() *DeleteAcmeDnsProviderReq {
	return &DeleteAcmeDnsProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAcmeDnsProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAcmeDnsProviderResp struct {
	Meta *basedto.Meta `json:"meta"`
}
