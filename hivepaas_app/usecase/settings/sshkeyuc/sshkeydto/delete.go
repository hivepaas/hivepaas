package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSSHKeyReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSSHKeyReq() *DeleteSSHKeyReq {
	return &DeleteSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSSHKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSSHKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
}
