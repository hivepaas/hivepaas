package sslcertdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSSLCertReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSSLCertReq() *DeleteSSLCertReq {
	return &DeleteSSLCertReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSSLCertReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSSLCertResp struct {
	Meta *basedto.Meta `json:"meta"`
}
