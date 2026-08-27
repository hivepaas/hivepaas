package sslproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateSSLProviderReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSLProviderResp struct {
	Meta *basedto.Meta `json:"meta"`
}
