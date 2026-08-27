package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSSHKeyReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSSHKeyReq() *DeleteSSHKeyReq {
	return &DeleteSSHKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSSHKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSSHKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
}
