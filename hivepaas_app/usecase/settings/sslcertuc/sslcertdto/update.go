package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSSLCertReq struct {
	settings.UpdateSettingReq
	*SSLCertBaseReq
}

func NewUpdateSSLCertReq() *UpdateSSLCertReq {
	return &UpdateSSLCertReq{}
}

func (req *UpdateSSLCertReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
}
