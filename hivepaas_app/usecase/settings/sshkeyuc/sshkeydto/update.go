package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateSSHKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSHKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
}
