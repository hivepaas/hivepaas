package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSSHKeyReq struct {
	settings.UpdateSettingReq
	*SSHKeyBaseReq
}

func NewUpdateSSHKeyReq() *UpdateSSHKeyReq {
	return &UpdateSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSHKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSHKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
}
