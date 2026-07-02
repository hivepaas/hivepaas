package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSSLProviderReq struct {
	settings.UpdateSettingReq
	*SSLProviderBaseReq
}

func NewUpdateSSLProviderReq() *UpdateSSLProviderReq {
	return &UpdateSSLProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSLProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSLProviderResp struct {
	Meta *basedto.Meta `json:"meta"`
}
