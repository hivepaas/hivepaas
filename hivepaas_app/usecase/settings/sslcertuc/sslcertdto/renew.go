package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type RenewSSLCertReq struct {
	settings.GetSettingReq
}

func NewRenewSSLCertReq() *RenewSSLCertReq {
	return &RenewSSLCertReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RenewSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type RenewSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
}
