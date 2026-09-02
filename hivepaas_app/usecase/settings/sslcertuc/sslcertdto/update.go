package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateSSLCertReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
}
