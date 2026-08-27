package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RemoveMFATotpReq struct {
	Passcode string `json:"passcode"`
}

func NewRemoveMFATotpReq() *RemoveMFATotpReq {
	return &RemoveMFATotpReq{}
}

func (req *RemoveMFATotpReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Passcode, true,
		minPasscodeLen, maxPasscodeLen, "passcode")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RemoveMFATotpResp struct {
	Meta *basedto.Meta `json:"meta"`
}
