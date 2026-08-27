package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type RenewSSLCertReq struct {
	settings.GetSettingReq
}

func NewRenewSSLCertReq() *RenewSSLCertReq {
	return &RenewSSLCertReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RenewSSLCertReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RenewSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
}
