package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSSLProviderReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSSLProviderReq() *DeleteSSLProviderReq {
	return &DeleteSSLProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSSLProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSSLProviderResp struct {
	Meta *basedto.Meta `json:"meta"`
}
